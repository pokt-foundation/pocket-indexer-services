package main

import (
	"flag"
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

	runPlayground := flag.Bool("p", false, "Flag to activate playground")

	flag.Parse()

	http.Handle("/query", srv)

	if *runPlayground {
		http.Handle("/", playground.Handler("GraphQL playground", "/query"))
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	}

	fmt.Printf("Indexer server running in port:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
