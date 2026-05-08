package main

import (
	"net/http"
	"os"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewServer() http.Handler {
	schemaBytes, err := os.ReadFile("schema.graphql")
	if err != nil {
		panic("could not read schema.graphql: " + err.Error())
	}

	schema := graphql.MustParseSchema(string(schemaBytes), &Resolver{},
		graphql.UseFieldResolvers(),
	)

	mux := http.NewServeMux()
	mux.Handle("/graphql", otelhttp.NewHandler(authMiddleware(&relay.Handler{Schema: schema}), "graphql"))
	mux.Handle("/playground", otelhttp.NewHandler(playgroundHandler(), "playground"))

	return corsMiddleware(mux)
}
