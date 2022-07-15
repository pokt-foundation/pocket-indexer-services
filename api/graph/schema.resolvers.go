package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/pokt-foundation/pocket-go/provider"
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

func (r *queryResolver) QueryBlocks(ctx context.Context, page *int, perPage *int, order *provider.Order) (*model.BlocksResponse, error) {
	options := &postgresdriver.ReadBlocksOptions{
		Page:    defaultPage,
		PerPage: defaultPerPage,
		Order:   defaultOrder,
	}

	if page != nil {
		options.Page = *page
	}
	if perPage != nil {
		options.PerPage = *perPage
	}
	if order != nil {
		options.Order = *order
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
		TotalCount: int(quantity),
		PageCount:  len(blocks),
		TotalPages: getTotalPages(int(quantity), options.PerPage),
	}, nil
}

func (r *queryResolver) QueryTransactionByHash(ctx context.Context, hash string) (*model.GraphQLTransaction, error) {
	transaction, err := r.Reader.ReadTransactionByHash(hash)
	if err != nil {
		return nil, err
	}

	return convertIndexerTransactionToGrapQLTransaction(transaction), nil
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
		Transactions: convertMultipleIndexerTransactionsToGrapQLTransactions(transactions),
		Page:         options.Page,
		TotalCount:   int(quantity),
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
		Transactions: convertMultipleIndexerTransactionsToGrapQLTransactions(transactions),
		Page:         options.Page,
		TotalCount:   int(quantity),
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
		Transactions: convertMultipleIndexerTransactionsToGrapQLTransactions(transactions),
		Page:         options.Page,
		TotalCount:   int(quantity),
		PageCount:    len(transactions),
		TotalPages:   getTotalPages(int(quantity), options.PerPage),
	}, nil
}

func (r *queryResolver) QueryAccountByAddress(ctx context.Context, address string, height *int) (*model.GraphQLAccount, error) {
	options := &postgresdriver.ReadAccountByAddressOptions{}

	if height != nil {
		options.Height = *height
	}

	account, err := r.Reader.ReadAccountByAddress(address, options)
	if err != nil {
		return nil, err
	}

	return convertIndexerAccountToGraphQLAccount(account), nil
}

func (r *queryResolver) QueryAccounts(ctx context.Context, height *int, page *int, perPage *int) (*model.AccountsResponse, error) {
	readOptions := &postgresdriver.ReadAccountsOptions{
		Page:    defaultPage,
		PerPage: defaultPerPage,
	}
	quantityOptions := &postgresdriver.GetAccountsQuantityOptions{}

	if height != nil {
		readOptions.Height = *height
		quantityOptions.Height = *height
	}

	if page != nil {
		readOptions.Page = *page
	}

	if perPage != nil {
		readOptions.PerPage = *perPage
	}

	accounts, err := r.Reader.ReadAccounts(readOptions)
	if err != nil {
		return nil, err
	}

	quantity, err := r.Reader.GetAccountsQuantity(quantityOptions)
	if err != nil {
		return nil, err
	}

	return &model.AccountsResponse{
		Accounts:   convertMultipleIndexerAccountToGraphQLAccount(accounts),
		Page:       readOptions.Page,
		TotalCount: int(quantity),
		PageCount:  len(accounts),
		TotalPages: getTotalPages(int(quantity), readOptions.PerPage),
	}, nil
}

func (r *queryResolver) QueryNodeByAddress(ctx context.Context, address string, height *int) (*model.GraphQLNode, error) {
	options := &postgresdriver.ReadNodeByAddressOptions{}

	if height != nil {
		options.Height = *height
	}

	node, err := r.Reader.ReadNodeByAddress(address, options)
	if err != nil {
		return nil, err
	}

	return convertIndexerNodeToGraphQLNode(node), nil
}

func (r *queryResolver) QueryNodes(ctx context.Context, height *int, page *int, perPage *int) (*model.NodesResponse, error) {
	readOptions := &postgresdriver.ReadNodesOptions{
		Page:    defaultPage,
		PerPage: defaultPerPage,
	}
	quantityOptions := &postgresdriver.GetNodesQuantityOptions{}

	if height != nil {
		readOptions.Height = *height
		quantityOptions.Height = *height
	}

	if page != nil {
		readOptions.Page = *page
	}

	if perPage != nil {
		readOptions.PerPage = *perPage
	}

	nodes, err := r.Reader.ReadNodes(readOptions)
	if err != nil {
		return nil, err
	}

	quantity, err := r.Reader.GetNodesQuantity(quantityOptions)
	if err != nil {
		return nil, err
	}

	return &model.NodesResponse{
		Nodes:      convertMultipleIndexerNodeToGraphQLNode(nodes),
		Page:       readOptions.Page,
		TotalCount: int(quantity),
		PageCount:  len(nodes),
		TotalPages: getTotalPages(int(quantity), readOptions.PerPage),
	}, nil
}

func (r *queryResolver) QueryAppByAddress(ctx context.Context, address string, height *int) (*model.GraphQLApp, error) {
	options := &postgresdriver.ReadAppByAddressOptions{}

	if height != nil {
		options.Height = *height
	}

	app, err := r.Reader.ReadAppByAddress(address, options)
	if err != nil {
		return nil, err
	}

	return convertIndexerAppToGraphQLApp(app), nil
}

func (r *queryResolver) QueryApps(ctx context.Context, height *int, page *int, perPage *int) (*model.AppsResponse, error) {
	readOptions := &postgresdriver.ReadAppsOptions{
		Page:    defaultPage,
		PerPage: defaultPerPage,
	}
	quantityOptions := &postgresdriver.GetAppsQuantityOptions{}

	if height != nil {
		readOptions.Height = *height
		quantityOptions.Height = *height
	}

	if page != nil {
		readOptions.Page = *page
	}

	if perPage != nil {
		readOptions.PerPage = *perPage
	}

	apps, err := r.Reader.ReadApps(readOptions)
	if err != nil {
		return nil, err
	}

	quantity, err := r.Reader.GetAppsQuantity(quantityOptions)
	if err != nil {
		return nil, err
	}

	return &model.AppsResponse{
		Apps:       convertMultipleIndexeraAppToGraphQLApp(apps),
		Page:       readOptions.Page,
		TotalCount: int(quantity),
		PageCount:  len(apps),
		TotalPages: getTotalPages(int(quantity), readOptions.PerPage),
	}, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
