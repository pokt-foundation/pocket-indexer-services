package graph

import indexerlib "github.com/pokt-foundation/pocket-indexer-lib"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// driver interface of needed functions for the db driver
type driver interface {
	ReadBlocks() ([]*indexerlib.Block, error)
	ReadTransactions() ([]*indexerlib.Transaction, error)
}

// Resolver struct handler for dependency injection to GraphQL operations
type Resolver struct {
	Driver driver
}
