package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	indexer "github.com/pokt-foundation/pocket-indexer-lib"
	postgresdriver "github.com/pokt-foundation/pocket-indexer-lib/postgres-driver"
	"github.com/pokt-foundation/pocket-indexer-services/api/graph/generated"
	"github.com/pokt-foundation/pocket-indexer-services/api/graph/model"
)

func (r *queryResolver) QueryBlockByHash(ctx context.Context, hash string) (*indexer.Block, error) {
	return r.Reader.ReadBlockByHash(hash)
}

func (r *queryResolver) QueryBlockByHeight(ctx context.Context, height int) (*indexer.Block, error) {
	return r.Reader.ReadBlockByHeight(height)
}

func (r *queryResolver) QueryBlocks(ctx context.Context, page *int, perPage *int) (*model.BlocksResponse, error) {
	options := &postgresdriver.ReadBlocksOptions{
		Page:    defaultPage,
		PerPage: defaultPerPage,
	}

	if page != nil {
		options.Page = *page
	}

	if perPage != nil {
		options.PerPage = *perPage
	}

	blocks, err := r.Reader.ReadBlocks(options)
	if err != nil {
		return nil, err
	}

	quantity, err := r.Reader.GetBlocksQuantity()
	if err != nil {
		return nil, err
	}

	return &model.BlocksResponse{
		Blocks:     blocks,
		Page:       options.Page,
		PageCount:  len(blocks),
		TotalPages: getTotalPages(int(quantity), options.PerPage),
	}, nil
}

func (r *queryResolver) QueryTransactionByHash(ctx context.Context, hash string) (*indexer.Transaction, error) {
	return r.Reader.ReadTransactionByHash(hash)
}

func (r *queryResolver) QueryTransactionsByHeight(ctx context.Context, height int, page *int, perPage *int) (*model.TransactionsResponse, error) {
	options := &postgresdriver.ReadTransactionsByHeightOptions{
		Page:    defaultPage,
		PerPage: defaultPerPage,
	}

	if page != nil {
		options.Page = *page
	}

	if perPage != nil {
		options.PerPage = *perPage
	}

	transactions, err := r.Reader.ReadTransactionsByHeight(height, options)
	if err != nil {
		return nil, err
	}

	quantity, err := r.Reader.GetTransactionsQuantityByHeight(height)
	if err != nil {
		return nil, err
	}

	return &model.TransactionsResponse{
		Transactions: transactions,
		Page:         options.Page,
		PageCount:    len(transactions),
		TotalPages:   getTotalPages(int(quantity), options.PerPage),
	}, nil
}

func (r *queryResolver) QueryTransactions(ctx context.Context, page *int, perPage *int) (*model.TransactionsResponse, error) {
	options := &postgresdriver.ReadTransactionsOptions{
		Page:    defaultPage,
		PerPage: defaultPerPage,
	}

	if page != nil {
		options.Page = *page
	}

	if perPage != nil {
		options.PerPage = *perPage
	}

	transactions, err := r.Reader.ReadTransactions(options)
	if err != nil {
		return nil, err
	}

	quantity, err := r.Reader.GetTransactionsQuantity()
	if err != nil {
		return nil, err
	}

	return &model.TransactionsResponse{
		Transactions: transactions,
		Page:         options.Page,
		PageCount:    len(transactions),
		TotalPages:   getTotalPages(int(quantity), options.PerPage),
	}, nil
}

func (r *queryResolver) QueryTransactionsByAddress(ctx context.Context, address string, page *int, perPage *int) (*model.TransactionsResponse, error) {
	options := &postgresdriver.ReadTransactionsByAddressOptions{
		Page:    defaultPage,
		PerPage: defaultPerPage,
	}

	if page != nil {
		options.Page = *page
	}

	if perPage != nil {
		options.PerPage = *perPage
	}

	transactions, err := r.Reader.ReadTransactionsByAddress(address, options)
	if err != nil {
		return nil, err
	}

	quantity, err := r.Reader.GetTransactionsQuantityByAddress(address)
	if err != nil {
		return nil, err
	}

	return &model.TransactionsResponse{
		Transactions: transactions,
		Page:         options.Page,
		PageCount:    len(transactions),
		TotalPages:   getTotalPages(int(quantity), options.PerPage),
	}, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
