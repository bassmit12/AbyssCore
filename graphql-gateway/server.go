package main

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
)

// NewServer builds the GraphQL HTTP handler with WebSocket subscription support.
func NewServer() http.Handler {
	// TODO: wire gqlgen ExecutableSchema once codegen runs (Phase 3 final step)
	// For now returns a placeholder - replace with:
	//   srv := handler.New(NewExecutableSchema(Config{Resolvers: &Resolver{}}))
	srv := handler.NewDefaultServer(nil)

	// WebSocket transport for subscriptions
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	})

	mux := http.NewServeMux()
	mux.Handle("/graphql", authMiddleware(srv))
	mux.Handle("/playground", playgroundHandler())

	return mux
}

func playgroundHandler() http.Handler {
	return playground.Handler("AbyssCore GraphQL", "/graphql")
}
