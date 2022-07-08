package graph

import (
	"strconv"

	indexerlib "github.com/pokt-foundation/pocket-indexer-lib"
	postgresdriver "github.com/pokt-foundation/pocket-indexer-lib/postgres-driver"
	"github.com/pokt-foundation/pocket-indexer-services/api/graph/model"
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
	ReadAccountByAddress(address string, options *postgresdriver.ReadAccountByAddressOptions) (*indexerlib.Account, error)
	ReadAccounts(options *postgresdriver.ReadAccountsOptions) ([]*indexerlib.Account, error)
	GetAccountsQuantity(options *postgresdriver.GetAccountsQuantityOptions) (int64, error)
	ReadNodeByAddress(address string, options *postgresdriver.ReadNodeByAddressOptions) (*indexerlib.Node, error)
	ReadNodes(options *postgresdriver.ReadNodesOptions) ([]*indexerlib.Node, error)
	GetNodesQuantity(options *postgresdriver.GetNodesQuantityOptions) (int64, error)
	ReadAppByAddress(address string, options *postgresdriver.ReadAppByAddressOptions) (*indexerlib.App, error)
	ReadApps(options *postgresdriver.ReadAppsOptions) ([]*indexerlib.App, error)
	GetAppsQuantity(options *postgresdriver.GetAppsQuantityOptions) (int64, error)
}

func getTotalPages(quantity, perPage int) int {
	fullPages := quantity / perPage

	remainder := quantity % perPage
	if remainder > 0 {
		return fullPages + 1
	}

	return fullPages
}

func convertIndexerTransactionToGrapQLTransaction(transaction *indexerlib.Transaction) *model.GraphQLTransaction {
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

func convertMultipleIndexerTransactionsToGrapQLTransactions(transactions []*indexerlib.Transaction) []*model.GraphQLTransaction {
	graphqlTransactions := []*model.GraphQLTransaction{}

	for _, transaction := range transactions {
		graphqlTransactions = append(graphqlTransactions, convertIndexerTransactionToGrapQLTransaction(transaction))
	}

	return graphqlTransactions
}

func convertIndexerAccountToGraphQLAccount(account *indexerlib.Account) *model.GraphQLAccount {
	return &model.GraphQLAccount{
		Address:             account.Address,
		Height:              account.Height,
		Balance:             account.Balance.String(),
		BalanceDenomination: account.BalanceDenomination,
	}
}

func convertMultipleIndexerAccountToGraphQLAccount(accounts []*indexerlib.Account) []*model.GraphQLAccount {
	graphqlAccounts := []*model.GraphQLAccount{}

	for _, account := range accounts {
		graphqlAccounts = append(graphqlAccounts, convertIndexerAccountToGraphQLAccount(account))
	}

	return graphqlAccounts
}

func convertIndexerNodeToGraphQLNode(node *indexerlib.Node) *model.GraphQLNode {
	return &model.GraphQLNode{
		Address:    node.Address,
		Height:     node.Height,
		Jailed:     node.Jailed,
		PublicKey:  node.PublicKey,
		ServiceURL: node.ServiceURL,
		Tokens:     node.Tokens.String(),
	}
}

func convertMultipleIndexerNodeToGraphQLNode(nodes []*indexerlib.Node) []*model.GraphQLNode {
	graphqlNodes := []*model.GraphQLNode{}

	for _, node := range nodes {
		graphqlNodes = append(graphqlNodes, convertIndexerNodeToGraphQLNode(node))
	}

	return graphqlNodes
}

func convertIndexerAppToGraphQLApp(app *indexerlib.App) *model.GraphQLApp {
	return &model.GraphQLApp{
		Address:      app.Address,
		Height:       app.Height,
		Jailed:       app.Jailed,
		PublicKey:    app.PublicKey,
		StakedTokens: app.StakedTokens.String(),
	}
}

func convertMultipleIndexeraAppToGraphQLApp(apps []*indexerlib.App) []*model.GraphQLApp {
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
