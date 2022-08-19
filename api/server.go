// Package main runs the GraphQL API service for the indexer
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
	"github.com/pokt-foundation/utils-go/environment"
)

var (
	connectionString = environment.GetString("CONNECTION_STRING", "")
	port             = environment.GetString("PORT", "8080")
	runPlayground    = environment.GetBool("RUN_PLAYGROUND", true)
)

func healthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Pocket Indexer API is up and running!"))
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	driver, err := postgresdriver.NewPostgresDriverFromConnectionString(connectionString)
	if err != nil {
		panic(fmt.Sprintf("connection to database failed with error: %s", err.Error()))
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		Reader: driver,
	}}))

	http.Handle("/", healthCheck())
	http.Handle("/query", srv)

	if runPlayground {
		http.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
		log.Printf("connect to http://localhost:%s/playground for GraphQL playground", port)
	}

	log.Printf("Indexer server running in port:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
