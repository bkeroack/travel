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

// Options for Travel router
// "Strict" means to follow Pyramid traversal semanics -- handler name can only be "" or the latest token in path when root_tree lookup
// failed (everything beyond that is the subpath).
// Non-strict (permissive) means that the handler name is always the latest token in the path (regardless if lookup fully succeeds).
// In both instances the default behavior can be overridden by having an explicit handler name token within the root tree node.
type TravelOptions struct {
	SubpathMaxLength map[string]int // Map of method verb to subpath length limit for requests of that type
	StrictTraversal  bool           // Obey Pyramid traversal semantics (do not enforce subpath limits, use handler names from path only)
}

// Request context passed to request handler
type Context struct {
	RootTree   map[string]interface{} // Root tree as processed by this request (thread-local)
	CurrentObj interface{}            // Current object from root tree
	Path       []string               // tokenized URL path
	options    *TravelOptions         // Options passed to router
	Subpath    []string               // Tokenized subpath for this request
}

// Travel router
type Router struct {
	rtf     RootTreeFunc
	hm      HandlerMap
	eh      TravelErrorHandler
	tokens  []string
	options *TravelOptions
}

// Result of running traversal algorithm
type TraversalResult struct {
	h  string      // handler name
	co interface{} // current object
	sp []string    // tokenized subpath
}

// Create a new Travel router. Parameters: callback function to fetch root tree, map of handler names to functions,
// default request error handler, options
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

// Fetch the root tree, re-run traversal and update Context fields.
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

// Walk back n nodes in tokenized path, return root tree object at that node.
func (c *Context) WalkBack(n uint) (map[string]interface{}, error) {
	new_path := c.Path[0 : len(c.Path)-int(n)]
	if len(new_path) == 0 {
		new_path = []string{""}
	}
	tr, err := doTraversal(c.RootTree, new_path, 0, c.options.StrictTraversal)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return tr.co.(map[string]interface{}), nil
}

func doTraversal(rt map[string]interface{}, tokens []string, spl int, strict bool) (TraversalResult, TraversalError) {
	var cur_obj interface{}
	var ok bool

	get_hn := func(token string, found bool) string {
		if (strict && found) || !strict {
			return ""
		} else {
			return token
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
							return TraversalResult{ // last token, token lookup success, cur_obj is map, explicit handler found
								h:  hns,
								co: co2,
								sp: []string{},
							}, nil
						} else {
							return TraversalResult{ // last token, token lookup success, cur_obj is map, no handler key
								h:  get_hn(t, true),
								co: co2,
								sp: []string{},
							}, nil
						}
					default:
						return TraversalResult{ // last token, token lookup success, cur_obj is not a map
							h:  get_hn(t, true),
							co: cur_obj,
							sp: []string{},
						}, nil
					}
				} // next iteration
			} else {
				// not found
				sp := tokens[i:len(tokens)]
				if len(sp) <= spl || len(tokens) == 1 {
					return TraversalResult{ // token not found, subpath_limit not exceeded
						h:  get_hn(t, false),
						co: co,
						sp: sp,
					}, nil
				} else {
					return TraversalResult{}, NotFoundError(tokens) // token not found, subpath limit exceeded
				}
			}
		default:
			if i == len(tokens)-1 {
				return TraversalResult{ // last token, current object is not a map
					h:  "",
					co: cur_obj,
					sp: []string{},
				}, nil
			} else {
				return TraversalResult{ // tokens remaining but cur_obj is not a map so traversal cannot continue
					h:  get_hn(t, false),
					co: cur_obj,
					sp: tokens[i : len(tokens)-1],
				}, nil
			}
		}
	}
	return TraversalResult{}, InternalError("received empty path")
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
	if terr != nil {
		r.eh(w, req, terr)
		return
	}
	log.Printf("got handler name: %v\n", tr.h)
	if h, ok := r.hm[tr.h]; ok {
		c := Context{
			RootTree:   rt,
			CurrentObj: tr.co,
			Path:       r.tokens,
			Subpath:    tr.sp,
			options:    r.options,
		}
		h(w, req, &c)
	} else {
		r.eh(w, req, UnknownHandlerError(r.tokens))
	}
}
