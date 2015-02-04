package travel

import (
	"fmt"
)

type TraversalNotFoundError []string
type TraversalRootTreeError string
type TraversalInternalError string
type TraversalIllegalSubpath []string

func (t TraversalNotFoundError) Error() string {
	return fmt.Sprintf("404 Not Found: %v".format(t))
}

func (t TraversalRootTreeError) Error() string {
	return fmt.Sprintf("Error retrieving root_tree: %v".format(t))
}

func (t TraversalInternalError) Error() string {
	return fmt.Sprintf("Internal traversal error (bug?): %v".format(t))
}

func (t TraversalIllegalSubpath) Error() string {
	return fmt.Sprintf("Subpath exceeded allowed length: %v".format(t))
}

func (t TraversalNotFoundError) New(path []string) TraversalNotFoundError {
	return path
}
