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
type TravelErrorHandler func(http.ResponseWriter, *http.Request, TraversalError)
type RootTreeFunc func() (map[string]interface{}, error)
type HandlerMap map[string]TravelHandler

type TravelOptions struct {
	SubpathMaxLength map[string]int
	StrictTraversal  bool
}

type Context struct {
	RootTree   map[string]interface{}
	CurrentObj interface{}
	Path       []string
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

	tr, err := doTraversal(rt, c.Path, spl, c.options.StrictTraversal)
	if err != nil {
		return err
	}
	c.CurrentObj = tr.co
	c.RootTree = rt
	c.Subpath = tr.sp
	return nil
}

func (c *Context) WalkBack(n uint) map[string]interface{} {
	if int(n) > len(c.Path) {
		return map[string]interface{}{}
	}
	ti := len(c.Path) - int(n)
	var co map[string]interface{}
	for i := 0; i <= ti; i++ {
		co = c.RootTree[c.Path[i]].(map[string]interface{})
	}
	return co
}

func doTraversal(rt map[string]interface{}, tokens []string, spl int, strict bool) (TraversalResult, TraversalError) {
	var cur_obj interface{}
	var ok bool

	get_hn := func(i int, l bool) string {
		if l {
			if strict {
				return ""
			} else {
				return tokens[i]
			}
		}
		if strict || len(tokens) == 1 {
			return tokens[i]
		} else {
			return tokens[i-1]
		}
	}

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
								h:  get_hn(i, true),
								co: co2,
								sp: []string{},
							}, nil
						}
					default:
						return TraversalResult{
							h:  get_hn(i, true),
							co: cur_obj,
							sp: []string{},
						}, nil
					}
				} // next iteration
			} else {
				// not found
				sp := tokens[i:len(tokens)]
				log.Printf("len(sp): %v, spl: %v, i: %v, len(tokens): %v", len(sp), spl, i, len(tokens))
				if len(sp) <= spl {
					var hn string
					if len(sp) == len(tokens) {
						hn = ""
					} else {
						hn = get_hn(i, false)
					}
					return TraversalResult{
						h:  hn,
						co: co,
						sp: sp,
					}, nil
				} else {
					return TraversalResult{}, NotFoundError(tokens)
				}
			}
		default:
			log.Printf("default")
			if i == len(tokens)-1 {
				return TraversalResult{
					h:  "",
					co: cur_obj,
					sp: []string{},
				}, nil
			} else {
				return TraversalResult{
					h:  get_hn(i, false),
					co: cur_obj,
					sp: tokens[i : len(tokens)-1],
				}, nil
			}
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
		return
	}

	var spl int
	if v, ok := r.options.SubpathMaxLength[req.Method]; ok {
		spl = v
	} else {
		spl = 0
	}

	tr, terr := doTraversal(rt, r.tokens, spl, r.options.StrictTraversal)
	log.Printf("got handler name: %v\n", tr.h)
	if terr != nil {
		r.eh(w, req, terr)
		return
	}
	if h, ok := r.hm[tr.h]; ok {
		c := Context{
			RootTree:   rt,
			CurrentObj: tr.co,
			Path:       r.tokens,
			Subpath:    tr.sp,
		}
		h(w, req, &c)
	} else {
		r.eh(w, req, UnknownHandlerError(r.tokens))
	}
}
