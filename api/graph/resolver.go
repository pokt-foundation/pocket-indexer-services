package graph

import (
	indexerlib "github.com/pokt-foundation/pocket-indexer-lib"
	postgresdriver "github.com/pokt-foundation/pocket-indexer-lib/postgres-driver"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// reader interface of needed functions for the db reader
type reader interface {
	ReadTransactions(options *postgresdriver.ReadTransactionsOptions) ([]*indexerlib.Transaction, error)
	ReadTransactionsByAddress(address string, options *postgresdriver.ReadTransactionsByAddressOptions) ([]*indexerlib.Transaction, error)
	ReadTransaction(hash string) (*indexerlib.Transaction, error)
	ReadBlocks(options *postgresdriver.ReadBlocksOptions) ([]*indexerlib.Block, error)
	ReadBlock(hash string) (*indexerlib.Block, error)
}

// Resolver struct handler for dependency injection to GraphQL operations
type Resolver struct {
	Reader reader
}
