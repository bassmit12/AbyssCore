package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/playground"
)

func playgroundHandler() http.Handler {
	return playground.Handler("AbyssCore GraphQL", "/graphql")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4001"
	}

	srv := NewServer()

	log.Printf("GraphQL gateway listening on :%s", port)
	log.Printf("Playground: http://localhost:%s/playground", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
