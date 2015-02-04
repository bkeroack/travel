package travel

import (
	"fmt"
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

type Context struct {
	RootTree   map[string]interface{}
	CurrentObj interface{}
	tokens     []string
	Subpath    []string
}

type Router struct {
	rtf    RootTreeFunc
	hm     HandlerMap
	eh     TravelErrorHandler
	tokens []string
}

type TraversalResult struct {
	h  string
	co interface{}
	sp []string
}

type TravelOptions struct {
	SubpathMaxLength map[string]int
}

func NewRouter(rtf RootTreeFunc, hm HandlerMap, eh TravelErrorHandler) *Router {
	return &Router{
		rtf: rtf,
		hm:  hm,
		eh:  eh,
	}
}

func (c *Context) Refresh(rtf RootTreeFunc) error {
	rt, err := rtf()
	if err != nil {
		return err
	}
	tr, err := doTraversal(rt, c.tokens)
	if err != nil {
		return err
	}
	c.CurrentObj = tr.co
	c.RootTree = rt
	c.Subpath = tr.sp
	return nil
}

func doTraversal(rt map[string]interface{}, tokens []string) (TraversalResult, error) {
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
				return TraversalResult{}, fmt.Errorf("404 Not Found")
			}
		default:
			return TraversalResult{
				h:  t,
				co: cur_obj,
				sp: tokens[i : len(tokens)-1],
			}, nil
		}
	}
	return TraversalResult{}, fmt.Errorf("traversal never completed")
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if req.URL.Path[0] == '/' {
		req.URL.Path = strings.TrimLeft(req.URL.Path, "/")
	}
	if req.URL.Path[len(req.URL.Path)-1] == '/' {
		req.URL.Path = strings.TrimRight(req.URL.Path, "/")
	}
	r.tokens = strings.Split(req.URL.Path, "/")

	rt, err := r.rtf()
	if err != nil {
		r.eh(w, req, fmt.Errorf("error getting root_tree"))
	}
	tr, err := doTraversal(rt, r.tokens)
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
		r.eh(w, req, fmt.Errorf("handler not found"))
	}
}
