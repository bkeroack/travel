package travel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"
)

var test_response []byte

func load_roottree(d []byte) (map[string]interface{}, error) {
	var rt interface{}
	err := json.Unmarshal(d, &rt)
	if err != nil {
		return map[string]interface{}{}, err
	}

	switch v := rt.(type) {
	case map[string]interface{}:
		return v, nil
	default:
		return map[string]interface{}{}, fmt.Errorf("incorrect json in root tree")
	}
}

func new_request(verb string, url string) *http.Request {
	r, err := http.NewRequest(verb, url, nil)
	if err != nil {
		log.Fatalf("error creating new request: %v\n", err)
	}
	return r
}

type test_response_writer struct {
	hh     http.Header
	status int
}

func (tw test_response_writer) Header() http.Header {
	return http.Header(tw.hh)
}

func (tw test_response_writer) Write(b []byte) (int, error) {
	test_response = b
	return len(b), nil
}

func (tw test_response_writer) WriteHeader(c int) {
	tw.status = c
}

func new_responsewriter() test_response_writer {
	return test_response_writer{
		hh:     make(http.Header),
		status: 0,
	}
}

func unmarshal_response() (map[string]interface{}, error) {
	var d interface{}
	err := json.Unmarshal(test_response, &d)
	if err != nil {
		return map[string]interface{}{}, err
	}
	switch v := d.(type) {
	case map[string]interface{}:
		return v, nil
	default:
		return map[string]interface{}{}, fmt.Errorf("incorrect json in response")
	}
}

func test_handler(w http.ResponseWriter, r *http.Request, c *Context, v string) {
	resp := map[string]interface{}{
		"resp":    v,
		"context": c.CurrentObj,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error marshalling json response: %v\n", err)
	}
	w.Write(b)
}

func test_error_handler(w http.ResponseWriter, r *http.Request, err TraversalError) {
	log.Printf("test_error_handler called\n")
	test_handler(w, r, &Context{}, err.Error())
}

func test_request(r *Router, v string, p string) map[string]interface{} {
	req := new_request(v, p)
	rw := new_responsewriter()
	r.ServeHTTP(rw, req)
	resp, err := unmarshal_response()
	if err != nil {
		return map[string]interface{}{}
	}
	return resp
}

func TestSimpleTraversal(t *testing.T) {
	rtf := func() (map[string]interface{}, error) {
		rt := `
		{
			"foo": {
				"bar": {
					"baz": {},
					"%handler": "bar"
				}
			}
		}`
		return load_roottree([]byte(rt))
	}

	var bar_handler TravelHandler
	bar_handler = func(w http.ResponseWriter, r *http.Request, c *Context) {
		test_handler(w, r, c, "bar")
	}

	baz_handler := func(w http.ResponseWriter, r *http.Request, c *Context) {
		test_handler(w, r, c, "baz")
	}

	def_handler := func(w http.ResponseWriter, r *http.Request, c *Context) {
		test_handler(w, r, c, "default")
	}

	hm := map[string]TravelHandler{
		"bar": bar_handler,
		"baz": baz_handler,
		"":    def_handler,
	}

	o := TravelOptions{
		StrictTraversal: true,
	}

	r, err := NewRouter(rtf, hm, test_error_handler, &o)
	if err != nil {
		t.Errorf("NewRouter error: %v\n", err)
	}
	resp := test_request(r, "GET", "/foo/bar")
	if resp["resp"] != "bar" {
		t.Errorf("Incorrect response: %v\n", resp)
		return
	}

	resp = test_request(r, "GET", "/foo/bar/baz")
	if resp["resp"] != "default" {
		t.Errorf("Incorrect response: %v\n", resp)
		return
	}
}

func TestPermissiveTraversal(t *testing.T) {
	rtf := func() (map[string]interface{}, error) {
		rt := `
		{
			"accounts": {
				"users": {
					"mary": {
						"%handler": "user"
					}
				}
			}
		}`
		return load_roottree([]byte(rt))
	}

	accounts_handler := func(w http.ResponseWriter, r *http.Request, c *Context) {
		test_handler(w, r, c, "accounts")
	}

	users_handler := func(w http.ResponseWriter, r *http.Request, c *Context) {
		test_handler(w, r, c, "users")
	}

	user_handler := func(w http.ResponseWriter, r *http.Request, c *Context) {
		test_handler(w, r, c, "user")
	}

	hm := map[string]TravelHandler{
		"accounts": accounts_handler,
		"users":    users_handler,
		"user":     user_handler,
	}

	r, err := NewRouter(rtf, hm, test_error_handler, nil)
	if err != nil {
		t.Errorf("NewRouter error: %v\n", err)
	}
	resp := test_request(r, "GET", "/accounts")
	if resp["resp"] != "accounts" {
		t.Errorf("Incorrect response: %v\n", resp)
		return
	}

	resp = test_request(r, "GET", "/accounts/users")
	if resp["resp"] != "users" {
		t.Errorf("Incorrect response: %v\n", resp)
		return
	}

	resp = test_request(r, "GET", "/accounts/users/mary")
	if resp["resp"] != "user" {
		t.Errorf("Incorrect response: %v\n", resp)
		return
	}
}
