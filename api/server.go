package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	postgresdriver "github.com/pokt-foundation/pocket-indexer-lib/postgres-driver"
	"github.com/pokt-foundation/pocket-indexer-services/api/graph"
	"github.com/pokt-foundation/pocket-indexer-services/api/graph/generated"
	"github.com/pokt-foundation/pocket-indexer-services/pkg/environment"
)

var (
	connectionString = environment.GetString("CONNECTION_STRING", "")
	port             = environment.GetString("PORT", "8080")
)

func main() {
	driver, err := postgresdriver.NewPostgresDriverFromConnectionString(connectionString)
	if err != nil {
		panic(fmt.Sprintf("connection to database failed with error: %s", err.Error()))
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		Reader: driver,
	}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
