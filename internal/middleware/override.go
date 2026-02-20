package middleware

import "net/http"

// MethodOverrideMiddleware allows clients to override HTTP methods
func MethodOverrideMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for X-HTTP-Method-Override header
		if r.Method == http.MethodPost {
			if method := r.Header.Get("X-HTTP-Method-Override"); method != "" {
				r.Method = method
			}
		}
		next.ServeHTTP(w, r)
	})
}
