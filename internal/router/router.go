package router

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const paramsKey contextKey = "params"

// Router is our HTTP router
type Router struct {
	// Root node of our trie
	root *Node

	// NotFound handler for 404 responses
	notFound http.HandlerFunc

	middleware []Middleware
}

func New() *Router {
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
	r.root.insert(segments, method, handler, 0)
}

func (r *Router) Use(middleware ...Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	segments := splitPath(path)
	params := make(map[string]string)

	handler := r.findHandler(segments, r.root, req.Method, params)

	if handler == nil {
		if r.notFound != nil {
			handler = r.notFound
		} else {
			handler = http.NotFound
		}
	}

	// Store params in the request context
	if len(params) > 0 {
		ctx := context.WithValue(req.Context(), paramsKey, params)
		req = req.WithContext(ctx)
	}

	// Apply middleware in reverse order (last added, first executed)
	var h http.Handler = http.HandlerFunc(handler)
	for i := len(r.middleware) - 1; i >= 0; i-- {
		h = r.middleware[i](h)
	}

	h.ServeHTTP(w, req)
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

	// Static children are at the beginning of the children slice due to our sorting
	for _, child := range node.children {
		if child.isParam {
			// We've reached parameter nodes, no more static nodes to check
			break
		}

		if child.part == segment {
			if handler := r.findHandler(remainingSegments, child, method, params); handler != nil {
				return handler
			}
		}
	}

	// Then try param matches
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
