package main

import (
	"fmt"
	"net/http"

	"github.com/HilthonTT/http-go-router/internal/router"
)

func main() {
	r := router.NewRouter()

	r.GET("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})

	r.GET("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "List of users")
	})

	r.GET("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		params := router.Params(r)
		fmt.Fprintf(w, "User details for user: %s", params["id"])
	})

	r.POST("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Create a new user")
	})

	fmt.Println("Server starting on http://localhost:8080")

	http.ListenAndServe(":8080", r)
}
