package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type contextKey string

const paramsKey contextKey = "params"

// Node represents a node in the trie
type Node struct {
	// Part is the path segment this node represents
	part string

	// IsParam indicates if this node is a path parameter (like :id)
	isParam bool

	// Children contains child nodes
	children []*Node

	// Handlers stores handler funcs for different HTTP methods
	handlers map[string]http.HandlerFunc
}

// Router is our HTTP router
type Router struct {
	// Root node of our trie
	root *Node

	// NotFound handler for 404 responses
	notFound http.HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		root: &Node{
			part:     "",
			children: []*Node{},
			handlers: make(map[string]http.HandlerFunc),
		},
		notFound: http.NotFound,
	}
}

// GET registers a handler for GET requests
func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodGet, path, handler)
}

// POST registers a handler for POST requests
func (r *Router) POST(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodPost, path, handler)
}

// PUT registers a handler for PUT requests
func (r *Router) PUT(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodPut, path, handler)
}

// DELETE registers a handler for DELETE requests
func (r *Router) DELETE(path string, handler http.HandlerFunc) {
	r.Handle(http.MethodDelete, path, handler)
}

// NotFound sets the handler for 404 responses
func (r *Router) NotFound(handler http.HandlerFunc) {
	r.notFound = handler
}

// Handle registers a new handler for the given method and path
func (r *Router) Handle(method, path string, handler http.HandlerFunc) {
	if path[0] != '/' {
		panic("path must begin with '/'")
	}

	segments := splitPath(path)
	currentNode := r.root

	// Navigate through existing nodes as far as possible
	for i, segment := range segments {
		var nextNode *Node
		isParam := false

		// Check if this is a path parameter
		if len(segment) > 0 && segment[0] == ':' {
			isParam = true
			segment = segment[1:] // Remove the ':' prefix
		}

		// Look for an existing child node that matches this segment
		for _, child := range currentNode.children {
			if child.part == segment && child.isParam == isParam {
				nextNode = child
				break
			}
		}

		// If no matching child was found, create a new one
		if nextNode == nil {
			nextNode = &Node{
				part:     segment,
				isParam:  isParam,
				children: []*Node{},
				handlers: make(map[string]http.HandlerFunc),
			}
			currentNode.children = append(currentNode.children, nextNode)
		}

		currentNode = nextNode

		// If this is the last segment, add the handler
		if i == len(segments)-1 {
			if _, exists := currentNode.handlers[method]; exists {
				panic(fmt.Sprintf("handler already registered for %s %s", method, path))
			}
			currentNode.handlers[method] = handler
		}
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	segments := splitPath(path)
	params := make(map[string]string)

	handler := r.findHandler(segments, r.root, req.Method, params)

	if handler == nil {
		if r.notFound != nil {
			r.notFound(w, req)
		} else {
			http.NotFound(w, req)
		}
		return
	}

	// Store params in the request context
	if len(params) > 0 {
		ctx := context.WithValue(req.Context(), paramsKey, params)
		req = req.WithContext(ctx)
	}

	handler(w, req)
}

// findHandler recursively searches for a handler matching the given path segments
func (r *Router) findHandler(segments []string, node *Node, method string, params map[string]string) http.HandlerFunc {
	// If we've been processed all segments, check for a handler
	if len(segments) == 0 {
		if handler, ok := node.handlers[method]; ok {
			return handler
		}
		return nil
	}

	segment := segments[0]
	remainingSegments := segments[1:]

	// First try exact match
	for _, child := range node.children {
		if !child.isParam && child.part == segment {
			if handler := r.findHandler(remainingSegments, child, method, params); handler != nil {
				return handler
			}
		}
	}

	// Then try param match
	for _, child := range node.children {
		if child.isParam {
			// Store the param value
			params[child.part] = segment

			if handler := r.findHandler(remainingSegments, child, method, params); handler != nil {
				return handler
			}

			// Remove param if it didn't lead to a match
			delete(params, child.part)
		}
	}

	return nil
}

// Helper function to get URL parameters
func Params(r *http.Request) map[string]string {
	params, _ := r.Context().Value(paramsKey).(map[string]string)
	return params
}

// splitPath splits a path into segments
func splitPath(path string) []string {
	segments := strings.Split(path, "/")

	// Remove empty segments
	result := make([]string, 0)
	for _, s := range segments {
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}
