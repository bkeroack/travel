package travel

import (
	"fmt"
	"net/http"
)

type TraversalError interface {
	Error() string
	Code() int
}

type TraversalNotFoundError struct {
	path []string
	code int
}
type TraversalRootTreeError struct {
	err  error
	code int
}
type TraversalInternalError struct {
	msg  string
	code int
}
type TraversalUnknownHandlerError struct {
	path []string
	code int
}

func (t TraversalNotFoundError) Error() string {
	return fmt.Sprintf("404 Not Found: %v", t.path)
}

func (t TraversalNotFoundError) Code() int {
	return t.code
}

func (t TraversalUnknownHandlerError) Error() string {
	return fmt.Sprintf("Handler not found for route: %v\n", t.path)
}

func (t TraversalUnknownHandlerError) Code() int {
	return t.code
}

func (t TraversalRootTreeError) Error() string {
	return t.err.Error()
}

func (t TraversalRootTreeError) Code() int {
	return t.code
}

func (t TraversalInternalError) Error() string {
	return fmt.Sprintf("Internal traversal error (bug?): %v", t.msg)
}

func (t TraversalInternalError) Code() int {
	return t.code
}

func NotFoundError(r []string) TraversalError {
	return TraversalNotFoundError{
		path: r,
		code: http.StatusNotFound,
	}
}

func UnknownHandlerError(r []string) TraversalError {
	return TraversalUnknownHandlerError{
		path: r,
		code: http.StatusNotImplemented,
	}
}

func RootTreeError(err error) TraversalError {
	return TraversalRootTreeError{
		err:  err,
		code: http.StatusInternalServerError,
	}
}

func InternalError(m string) TraversalError {
	return TraversalInternalError{
		msg:  m,
		code: http.StatusInternalServerError,
	}
}
