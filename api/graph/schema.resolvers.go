package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	indexer "github.com/pokt-foundation/pocket-indexer-lib"
	"github.com/pokt-foundation/pocket-indexer-services/api/graph/generated"
)

// Blocks returns all blocks saved
func (r *queryResolver) Blocks(ctx context.Context) ([]*indexer.Block, error) {
	return r.Driver.ReadBlocks()
}

// Transactions returns all transactions saved
func (r *queryResolver) Transactions(ctx context.Context) ([]*indexer.Transaction, error) {
	return r.Driver.ReadTransactions()
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
