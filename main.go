package main

import (
	"fmt"
	"net/http"

	"github.com/HilthonTT/http-go-router/internal/middleware"
	"github.com/HilthonTT/http-go-router/internal/router"
)

func main() {
	r := router.New()
	r.Use(middleware.Logger)

	r.GET("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})

	api := r.Group("/api")

	api.GET("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "List of users")
	})

	api.GET("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		params := router.Params(r)
		fmt.Fprintf(w, "User details for user: %s", params["id"])
	})

	api.POST("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Create a new user")
	})

	fmt.Println("Server starting on http://localhost:8080")

	http.ListenAndServe(":8080", r)
}
