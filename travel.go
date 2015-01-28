package travel

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

const (
	h_token  = "{handler}"
	err_name = "_error"
)

type Context interface{}
type TravelHandler func(http.ResponseWriter, *http.Request, Context)
type TravelErrorHandler func(http.ResponseWriter, *http.Request, string)
type RootTree map[string]interface{}
type RootTreeFunc func() (RootTree, error)
type HandlerMap map[string]TravelHandler

type Router struct {
	rtf RootTreeFunc
	hm  HandlerMap
	eh  TravelErrorHandler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rt, err := r.rtf()
	if err != nil {
		return r.eh(w, req, "error loading root tree")
	}

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
	cur_obj = rt
	for i := range tokens {
		t := tokens[i]
		v := reflect.ValueOf(cur_obj)
		if v.Kind() == reflect.Map {
			if cur_obj, ok = cur_obj[t]; ok {
				if i == (len(tokens) - 1) {
					if h, ok = r.hm[h_token]; ok {
						return h(w, req, cur_obj)
					} else {
						if h, ok = r.hm[""]; ok {
							return h(w, req, cur_obj)
						} else {
							return r.eh(w, req, "successful traversal but no matching handler found")
						}
					}
				}
			} else {
				return http.NotFoundHandler(w, req)
			}
		} else {
			if h, ok = r.hm[t]; ok {
				return h(w, req, cur_obj)
			} else {
				return r.eh(w, req, fmt.Sprintf("handler not found: %v\n", t))
			}
		}
	}
}
