package travel

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	h_token = "{handler}"
)

type TravelHandler func(http.ResponseWriter, *http.Request, interface{})
type TravelErrorHandler func(http.ResponseWriter, *http.Request, string)
type RootTreeFunc func() (map[string]interface{}, error)
type HandlerMap map[string]TravelHandler

type Router struct {
	rtf RootTreeFunc
	hm  HandlerMap
	eh  TravelErrorHandler
}

func NewRouter(rtf RootTreeFunc, hm HandlerMap, eh TravelErrorHandler) *Router {
	return &Router{
		rtf: rtf,
		hm:  hm,
		eh:  eh,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if req.URL.Path[0] == '/' {
		req.URL.Path = strings.TrimLeft(req.URL.Path, "/")
	}
	if req.URL.Path[len(req.URL.Path)-1] == '/' {
		req.URL.Path = strings.TrimRight(req.URL.Path, "/")
	}
	tokens := strings.Split(req.URL.Path, "/")

	var cur_obj interface{}
	var ok bool
	var h TravelHandler

	cur_obj, err := r.rtf()
	if err != nil {
		r.eh(w, req, "error loading root tree")
		return
	}
	for i := range tokens {
		t := tokens[i]
		switch co := cur_obj.(type) {
		case map[string]interface{}:
			if cur_obj, ok = co[t]; ok {
				if i == (len(tokens) - 1) {
					if h, ok = r.hm[h_token]; ok {
						h(w, req, cur_obj)
						return
					} else {
						if h, ok = r.hm[""]; ok {
							h(w, req, cur_obj)
							return
						} else {
							r.eh(w, req, "successful traversal but no matching handler found")
							return
						}
					}
				} // next iteration
			} else {
				http.NotFound(w, req)
				return
			}
		default:
			if h, ok = r.hm[t]; ok {
				h(w, req, cur_obj)
				return
			} else {
				r.eh(w, req, fmt.Sprintf("handler not found: %v\n", t))
				return
			}
		}
	}
}
