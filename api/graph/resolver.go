package graph

import (
	indexerlib "github.com/pokt-foundation/pocket-indexer-lib"
	postgresdriver "github.com/pokt-foundation/pocket-indexer-lib/postgres-driver"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

var (
	defaultPerPage = 1000
	defaultPage    = 1
)

// reader interface of needed functions for the db reader
type reader interface {
	ReadTransactions(options *postgresdriver.ReadTransactionsOptions) ([]*indexerlib.Transaction, error)
	GetTransactionsQuantity() (int64, error)
	ReadTransactionsByAddress(address string, options *postgresdriver.ReadTransactionsByAddressOptions) ([]*indexerlib.Transaction, error)
	GetTransactionsQuantityByAddress(address string) (int64, error)
	ReadTransactionsByHeight(height int, options *postgresdriver.ReadTransactionsByHeightOptions) ([]*indexerlib.Transaction, error)
	GetTransactionsQuantityByHeight(height int) (int64, error)
	ReadTransactionByHash(hash string) (*indexerlib.Transaction, error)
	ReadBlocks(options *postgresdriver.ReadBlocksOptions) ([]*indexerlib.Block, error)
	GetBlocksQuantity() (int64, error)
	ReadBlockByHash(hash string) (*indexerlib.Block, error)
	ReadBlockByHeight(height int) (*indexerlib.Block, error)
}

func getTotalPages(quantity, perPage int) int {
	fullPages := quantity / perPage

	remainder := quantity % perPage
	if remainder > 0 {
		return fullPages + 1
	}

	return fullPages
}

// Resolver struct handler for dependency injection to GraphQL operations
type Resolver struct {
	Reader reader
}
