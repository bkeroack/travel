package travel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

const (
	rt_file = "test/root.json"
)

func load_roottree(p string) (map[string]interface{}, error) {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return map[string]interface{}{}, err
	}

	var rt interface{}
	err = json.Unmarshal(b, &rt)
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
	hh      http.Header
	status  int
	content []byte
}

func (tw test_response_writer) Header() http.Header {
	return http.Header(tw.hh)
}

func (tw test_response_writer) Write(b []byte) (int, error) {
	tw.content = b
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

func unmarshal_response(rw test_response_writer) (map[string]interface{}, error) {
	var d interface{}
	err := json.Unmarshal(rw.content, &d)
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

func test_handler(w http.ResponseWriter, r *http.Request, c interface{}, v string) {
	resp := map[string]interface{}{
		"resp":    v,
		"context": c,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error marshalling json response: %v\n", err)
	}
	w.Write(b)
}

func test_error_handler(w http.ResponseWriter, r *http.Request, e string) {
	test_handler(w, r, "error", e)
}

func TestSimpleTraversal(t *testing.T) {
	rtf := func() (map[string]interface{}, error) {
		return load_roottree(rt_file)
	}

	var bar_handler TravelHandler
	bar_handler = func(w http.ResponseWriter, r *http.Request, c interface{}) {
		test_handler(w, r, c, "bar")
	}

	baz_handler := func(w http.ResponseWriter, r *http.Request, c interface{}) {
		test_handler(w, r, c, "baz")
	}

	hm := map[string]TravelHandler{
		"bar": bar_handler,
		"baz": baz_handler,
	}

	r := NewRouter(rtf, hm, test_error_handler)
	req := new_request("GET", "/foo/bar")
	rw := new_responsewriter()
	r.ServeHTTP(rw, req)
	resp, err := unmarshal_response(rw)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v\n", err)
	}
	if resp["resp"] != "bar" {
		t.Errorf("Incorrect response: %v\n", resp)
	}
}
