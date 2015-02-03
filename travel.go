package travel

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	h_token = "%handler"
)

type TravelHandler func(http.ResponseWriter, *http.Request, interface{})
type TravelErrorHandler func(http.ResponseWriter, *http.Request, error)
type RootTreeFunc func() (map[string]interface{}, error)
type HandlerMap map[string]TravelHandler

type Context struct {
	RootTree   map[string]interface{}
	CurrentObj interface{}
	tokens     []string
}

type Router struct {
	rtf    RootTreeFunc
	hm     HandlerMap
	eh     TravelErrorHandler
	tokens []string
}

func NewRouter(rtf RootTreeFunc, hm HandlerMap, eh TravelErrorHandler) *Router {
	return &Router{
		rtf: rtf,
		hm:  hm,
		eh:  eh,
	}
}

func (*Context) Refresh() {

}

func doTraversal(rt map[string]interface{}, tokens []string) (h string, co interface{}, err error) {
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
							return hns, cur_obj, nil

						} else {
							return "", cur_obj, nil
						}
					default:
						return "", cur_obj, nil
					}
				} // next iteration
			} else {
				return "", cur_obj, fmt.Errorf("404 Not Found")
			}
		default:
			return t, cur_obj, nil
		}
	}
	return "", cur_obj, fmt.Errorf("traversal never completed")
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
	hn, co, err := doTraversal(rt, r.tokens)
	if err != nil {
		r.eh(w, req, err)
		return
	}
	if h, ok := r.hm[hn]; ok {
		c := Context{
			RootTree:   rt,
			CurrentObj: co,
			tokens:     r.tokens,
		}
		h(w, req, c)
	} else {
		r.eh(w, req, fmt.Errorf("handler not found"))
	}
}
