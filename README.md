Traversal-like Dynamic HTTP Routing in Go
=================================

* For usage/details: https://godoc.org/github.com/bkeroack/travel
* For example usage: https://github.com/bkeroack/travel-examples

Travel is an HTTP router that provides dynamic routing functionality similar to the "traversal" system from the Pyramid web framework in Python.

For details on the original traversal system see: http://docs.pylonsproject.org/docs/pyramid/en/latest/narr/traversal.html

Simply put, traversal allows you to route HTTP requests by providing a nested ``map[string]interface{}`` object called the
"root tree". Request URLs are tokenized and recursive lookup is performed on the root tree object.

Example:

If the request URL is ``/foo/bar/baz/123``, it is tokenized to the following:

```json
   ["foo", "bar", "baz", "123"]
```

Then the equivalent of the following lookup is performed:

```go
   root_tree["foo"]["bar"]["baz"]["123"]
```

The object that results from this lookup is the "current object". If traversal succeeds, a named handler is then invoked (looked up via the "handler map" provided to the router), otherwise the router returns an appropriate error (404, etc).

For details on how lookup translates to handler names, see the godoc documentation linked above. Travel allows users to emulate traditional
traversal mechanics while also providing several ways to modify behavior, such as handler name overrides within the root tree object and limitations on subpath length.