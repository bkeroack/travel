Traversal-like HTTP Routing in Go
=================================

* For usage/details: https://godoc.org/github.com/bkeroack/travel
* For example usage: https://github.com/bkeroack/travel-examples

Travel is an HTTP router that provides routing similar to the "traversal" system from the Pyramid web framework in Python.

For details on the original traversal system please read: http://docs.pylonsproject.org/docs/pyramid/en/latest/narr/traversal.html

Simply put, traversal allows you to dynamically route HTTP requests by providing a nested map[string]interface{} object called the
"root tree" (Pyramid calls this the *resource tree*). Request URLs are tokenized and recursive lookup is performed on the root
tree object.

Example:

If the request URL is ``/foo/bar/baz/123``, it is tokenized to the following:

```json
   ["foo", "bar", "baz", "123"]
```

Then the equivalent of the following lookup is performed:

```go
   root_tree["foo"]["bar"]["baz"]["123"]
```

The object that results from this lookup is the "current object" (Pyramid calls this the "context"--travel provides a context object that contains the current object).

How this maps to handlers (in Pyramid terminology: "views") depends upon the options passed when creating the router. Under traditional
traversal, if the lookup fully succeeded (no missing key errors), the name of the handler would be the empty string ("") which is considered
the default handler. If the lookup failed at any point, the handler name would be the token that failed and any remainder of the URL would
be passed to the handler as the "subpath" (see original traversal documentation linked above for more details).

Travel provides several ways to modify these semantics. For details, see godoc documentation.