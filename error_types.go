package travel

import (
	"fmt"
)

type TraversalNotFoundError []string
type TraversalRootTreeError struct {
	err error
}
type TraversalInternalError struct {
	m string
}
type TraversalIllegalSubpath []string
type TraversalUnknownHandlerError []string

func (t TraversalNotFoundError) Error() string {
	return fmt.Sprintf("404 Not Found: %v", t)
}

func (t TraversalUnknownHandlerError) Error() string {
	return fmt.Sprintf("Handler not found for route: %v\n", t)
}

func (t TraversalRootTreeError) Error() string {
	return t.err.Error()
}

func (t TraversalInternalError) Error() string {
	return fmt.Sprintf("Internal traversal error (bug?): %v", t.m)
}

func (t TraversalIllegalSubpath) Error() string {
	return fmt.Sprintf("Subpath exceeded allowed length: %v", t)
}

func NotFoundError(r []string) TraversalNotFoundError {
	return r
}

func UnknownHandlerError(r []string) TraversalUnknownHandlerError {
	return r
}

func RootTreeError(err error) TraversalRootTreeError {
	return TraversalRootTreeError{
		err: err,
	}
}

func InternalError(m string) TraversalInternalError {
	return TraversalInternalError{
		m: m,
	}
}

func IllegalSubpath(r []string) TraversalIllegalSubpath {
	return r
}
