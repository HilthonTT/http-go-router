package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkRouterSimple(b *testing.B) {
	router := New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request) {})
	router.GET("/user/:id", func(w http.ResponseWriter, r *http.Request) {})

	req, _ := http.NewRequest("GET", "/user/123", nil)

	for b.Loop() {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
