package travel

import (
	"fmt"
)

type TraversalNotFoundError struct {
	Path []string
	Code int
}
type TraversalRootTreeError struct {
	Err  error
	Code int
}
type TraversalInternalError struct {
	Msg  string
	Code int
}
type TraversalIllegalSubpath struct {
	Path []string
	Code int
}
type TraversalUnknownHandlerError struct {
	Path []string
	Code int
}

func (t TraversalNotFoundError) Error() string {
	return fmt.Sprintf("404 Not Found: %v", t.Path)
}

func (t TraversalUnknownHandlerError) Error() string {
	return fmt.Sprintf("Handler not found for route: %v\n", t.Path)
}

func (t TraversalRootTreeError) Error() string {
	return t.Err.Error()
}

func (t TraversalInternalError) Error() string {
	return fmt.Sprintf("Internal traversal error (bug?): %v", t.Msg)
}

func (t TraversalIllegalSubpath) Error() string {
	return fmt.Sprintf("Subpath exceeded allowed length: %v", t.Path)
}

func NotFoundError(r []string) TraversalNotFoundError {
	return TraversalNotFoundError{
		Path: r,
		Code: 404,
	}
}

func UnknownHandlerError(r []string) TraversalUnknownHandlerError {
	return TraversalUnknownHandlerError{
		Path: r,
		Code: 401,
	}
}

func RootTreeError(err error) TraversalRootTreeError {
	return TraversalRootTreeError{
		Err:  err,
		Code: 500,
	}
}

func InternalError(m string) TraversalInternalError {
	return TraversalInternalError{
		Msg:  m,
		Code: 500,
	}
}

func IllegalSubpath(r []string) TraversalIllegalSubpath {
	return TraversalIllegalSubpath{
		Path: r,
		Code: 401,
	}
}
