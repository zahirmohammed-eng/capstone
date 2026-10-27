package graphql

import (
	"net/http"

	"github.com/graphql-go/handler"
)

func NewHTTPHandler(s *Server) http.Handler {
	return handler.New(&handler.Config{
		Schema:     &s.Schema,
		Pretty:     true,
		GraphiQL:   true,
		Playground: true,
	})
}
