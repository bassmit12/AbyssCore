package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	srv := NewServer()

	http.Handle("/graphql", srv)
	http.Handle("/playground", playgroundHandler())

	log.Printf("GraphQL gateway listening on :%s", port)
	log.Printf("Playground: http://localhost:%s/playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
