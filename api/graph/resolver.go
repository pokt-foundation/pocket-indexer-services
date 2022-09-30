package graph

import (
	"strconv"

	"github.com/pokt-foundation/pocket-indexer-lib/types"
	"github.com/pokt-foundation/pocket-indexer-services/api/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

var (
	defaultPerPage = 1000
	defaultPage    = 1
	defaultOrder   = types.DescendantOrder
)

// reader interface of needed functions for the db reader
type reader interface {
	ReadTransactions(options *types.ReadTransactionsOptions) ([]*types.Transaction, error)
	GetTransactionsQuantity() (int64, error)
	ReadTransactionsByAddress(address string, options *types.ReadTransactionsByAddressOptions) ([]*types.Transaction, error)
	GetTransactionsQuantityByAddress(address string) (int64, error)
	ReadTransactionsByHeight(height int, options *types.ReadTransactionsByHeightOptions) ([]*types.Transaction, error)
	GetTransactionsQuantityByHeight(height int) (int64, error)
	ReadTransactionByHash(hash string) (*types.Transaction, error)
	ReadBlocks(options *types.ReadBlocksOptions) ([]*types.Block, error)
	GetBlocksQuantity() (int64, error)
	ReadBlockByHash(hash string) (*types.Block, error)
	ReadBlockByHeight(height int) (*types.Block, error)
	ReadAccountByAddress(address string, options *types.ReadAccountByAddressOptions) (*types.Account, error)
	ReadAccounts(options *types.ReadAccountsOptions) ([]*types.Account, error)
	GetAccountsQuantity(options *types.GetAccountsQuantityOptions) (int64, error)
	ReadNodeByAddress(address string, options *types.ReadNodeByAddressOptions) (*types.Node, error)
	ReadNodes(options *types.ReadNodesOptions) ([]*types.Node, error)
	GetNodesQuantity(options *types.GetNodesQuantityOptions) (int64, error)
	ReadAppByAddress(address string, options *types.ReadAppByAddressOptions) (*types.App, error)
	ReadApps(options *types.ReadAppsOptions) ([]*types.App, error)
	GetAppsQuantity(options *types.GetAppsQuantityOptions) (int64, error)
}

func getTotalPages(quantity, perPage int) int {
	fullPages := quantity / perPage

	remainder := quantity % perPage
	if remainder > 0 {
		return fullPages + 1
	}

	return fullPages
}

func convertIndexerTransactionToGrapQLTransaction(transaction *types.Transaction) *model.GraphQLTransaction {
	return &model.GraphQLTransaction{
		Hash:            transaction.Hash,
		FromAddress:     transaction.FromAddress,
		ToAddress:       transaction.ToAddress,
		AppPubKey:       transaction.AppPubKey,
		Blockchains:     transaction.Blockchains,
		MessageType:     transaction.MessageType,
		Height:          transaction.Height,
		Index:           transaction.Index,
		StdTx:           transaction.StdTx,
		TxResult:        transaction.TxResult,
		Tx:              transaction.Tx,
		Entropy:         strconv.Itoa(transaction.Entropy),
		Fee:             transaction.Fee,
		FeeDenomination: transaction.FeeDenomination,
		Amount:          transaction.Amount.String(),
	}
}

func convertMultipleIndexerTransactionsToGrapQLTransactions(transactions []*types.Transaction) []*model.GraphQLTransaction {
	graphqlTransactions := []*model.GraphQLTransaction{}

	for _, transaction := range transactions {
		graphqlTransactions = append(graphqlTransactions, convertIndexerTransactionToGrapQLTransaction(transaction))
	}

	return graphqlTransactions
}

func convertIndexerAccountToGraphQLAccount(account *types.Account) *model.GraphQLAccount {
	return &model.GraphQLAccount{
		Address:             account.Address,
		Height:              account.Height,
		Balance:             account.Balance.String(),
		BalanceDenomination: account.BalanceDenomination,
	}
}

func convertMultipleIndexerAccountToGraphQLAccount(accounts []*types.Account) []*model.GraphQLAccount {
	graphqlAccounts := []*model.GraphQLAccount{}

	for _, account := range accounts {
		graphqlAccounts = append(graphqlAccounts, convertIndexerAccountToGraphQLAccount(account))
	}

	return graphqlAccounts
}

func convertIndexerNodeToGraphQLNode(node *types.Node) *model.GraphQLNode {
	return &model.GraphQLNode{
		Address:    node.Address,
		Height:     node.Height,
		Jailed:     node.Jailed,
		PublicKey:  node.PublicKey,
		ServiceURL: node.ServiceURL,
		Tokens:     node.Tokens.String(),
	}
}

func convertMultipleIndexerNodeToGraphQLNode(nodes []*types.Node) []*model.GraphQLNode {
	graphqlNodes := []*model.GraphQLNode{}

	for _, node := range nodes {
		graphqlNodes = append(graphqlNodes, convertIndexerNodeToGraphQLNode(node))
	}

	return graphqlNodes
}

func convertIndexerAppToGraphQLApp(app *types.App) *model.GraphQLApp {
	return &model.GraphQLApp{
		Address:      app.Address,
		Height:       app.Height,
		Jailed:       app.Jailed,
		PublicKey:    app.PublicKey,
		StakedTokens: app.StakedTokens.String(),
	}
}

func convertMultipleIndexeraAppToGraphQLApp(apps []*types.App) []*model.GraphQLApp {
	graphqlApps := []*model.GraphQLApp{}

	for _, app := range apps {
		graphqlApps = append(graphqlApps, convertIndexerAppToGraphQLApp(app))
	}

	return graphqlApps
}

// Resolver struct handler for dependency injection to GraphQL operations
type Resolver struct {
	Reader reader
}
