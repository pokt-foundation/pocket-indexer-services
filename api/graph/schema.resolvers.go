package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	indexer "github.com/pokt-foundation/pocket-indexer-lib"
	postgresdriver "github.com/pokt-foundation/pocket-indexer-lib/postgres-driver"
	"github.com/pokt-foundation/pocket-indexer-services/api/graph/generated"
)

func (r *queryResolver) QueryBlock(ctx context.Context, hash string) (*indexer.Block, error) {
	return r.Reader.ReadBlock(hash)
}

func (r *queryResolver) QueryBlocks(ctx context.Context, page *int, perPage *int) ([]*indexer.Block, error) {
	options := &postgresdriver.ReadBlocksOptions{}

	if page != nil {
		options.Page = *page
	}

	if perPage != nil {
		options.PerPage = *perPage
	}

	return r.Reader.ReadBlocks(options)
}

func (r *queryResolver) QueryTransaction(ctx context.Context, hash string) (*indexer.Transaction, error) {
	return r.Reader.ReadTransaction(hash)
}

func (r *queryResolver) QueryTransactions(ctx context.Context, page *int, perPage *int) ([]*indexer.Transaction, error) {
	options := &postgresdriver.ReadTransactionsOptions{}

	if page != nil {
		options.Page = *page
	}

	if perPage != nil {
		options.PerPage = *perPage
	}

	return r.Reader.ReadTransactions(options)
}

func (r *queryResolver) QueryTransactionsByAddress(ctx context.Context, address string, page *int, perPage *int) ([]*indexer.Transaction, error) {
	options := &postgresdriver.ReadTransactionsByAddressOptions{}

	if page != nil {
		options.Page = *page
	}

	if perPage != nil {
		options.PerPage = *perPage
	}

	return r.Reader.ReadTransactionsByAddress(address, options)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
