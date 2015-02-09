package travel

import (
	"log"
	"net/http"
	"strings"
)

const (
	h_token = "%handler"
)

type TravelHandler func(http.ResponseWriter, *http.Request, *Context)
type TravelErrorHandler func(http.ResponseWriter, *http.Request, error)
type RootTreeFunc func() (map[string]interface{}, error)
type HandlerMap map[string]TravelHandler

type TravelOptions struct {
	SubpathMaxLength map[string]int
}

type Context struct {
	RootTree   map[string]interface{}
	CurrentObj interface{}
	tokens     []string
	options    *TravelOptions
	Subpath    []string
}

type Router struct {
	rtf     RootTreeFunc
	hm      HandlerMap
	eh      TravelErrorHandler
	tokens  []string
	options *TravelOptions
}

type TraversalResult struct {
	h  string
	co interface{}
	sp []string
}

func NewRouter(rtf RootTreeFunc, hm HandlerMap, eh TravelErrorHandler, o *TravelOptions) *Router {
	if o == nil {
		o = &TravelOptions{
			SubpathMaxLength: map[string]int{},
		}
	}
	return &Router{
		rtf:     rtf,
		hm:      hm,
		eh:      eh,
		options: o,
	}
}

func (c *Context) Refresh(rtf RootTreeFunc, m string) error {
	rt, err := rtf()
	if err != nil {
		return RootTreeError(err)
	}

	var spl int
	if v, ok := c.options.SubpathMaxLength[m]; ok {
		spl = v
	} else {
		spl = 0
	}

	tr, err := doTraversal(rt, c.tokens, spl)
	if err != nil {
		return err
	}
	c.CurrentObj = tr.co
	c.RootTree = rt
	c.Subpath = tr.sp
	return nil
}

func doTraversal(rt map[string]interface{}, tokens []string, spl int) (TraversalResult, error) {
	var cur_obj interface{}
	var ok bool

	cur_obj = rt
	for i := range tokens {
		t := tokens[i]
		switch co := cur_obj.(type) {
		case map[string]interface{}:
			if cur_obj, ok = co[t]; ok {
				if i == len(tokens)-1 {
					switch co2 := cur_obj.(type) {
					case map[string]interface{}:
						if hn, ok := co2[h_token]; ok {
							hns := hn.(string)
							return TraversalResult{
								h:  hns,
								co: co2,
								sp: []string{},
							}, nil

						} else {
							return TraversalResult{
								h:  "",
								co: co2,
								sp: []string{},
							}, nil
						}
					default:
						return TraversalResult{
							h:  "",
							co: cur_obj,
							sp: []string{},
						}, nil
					}
				} // next iteration
			} else {
				// not found
				sp := tokens[i : len(tokens)-1]
				if len(sp) <= spl {
					return TraversalResult{
						h:  t,
						co: co,
						sp: sp,
					}, nil
				} else if len(sp) > spl {
					return TraversalResult{}, IllegalSubpath(tokens)
				} else {
					return TraversalResult{}, NotFoundError(tokens)
				}
			}
		default:
			return TraversalResult{
				h:  t,
				co: cur_obj,
				sp: tokens[i : len(tokens)-1],
			}, nil
		}
	}
	return TraversalResult{}, InternalError("traversal never completed")
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("got path: %v\n", req.URL.Path)
	if req.URL.Path[0] == '/' {
		req.URL.Path = strings.TrimLeft(req.URL.Path, "/")
	}
	if len(req.URL.Path) > 0 {
		if req.URL.Path[len(req.URL.Path)-1] == '/' {
			req.URL.Path = strings.TrimRight(req.URL.Path, "/")
		}
	}
	r.tokens = strings.Split(req.URL.Path, "/")

	rt, err := r.rtf()
	if err != nil {
		r.eh(w, req, RootTreeError(err))
	}

	var spl int
	if v, ok := r.options.SubpathMaxLength[req.Method]; ok {
		spl = v
	} else {
		spl = 0
	}

	tr, err := doTraversal(rt, r.tokens, spl)
	log.Printf("got handler name: %v\n", tr.h)
	if err != nil {
		r.eh(w, req, err)
		return
	}
	if h, ok := r.hm[tr.h]; ok {
		c := Context{
			RootTree:   rt,
			CurrentObj: tr.co,
			tokens:     r.tokens,
			Subpath:    tr.sp,
		}
		h(w, req, &c)
	} else {
		r.eh(w, req, UnknownHandlerError(r.tokens))
	}
}
