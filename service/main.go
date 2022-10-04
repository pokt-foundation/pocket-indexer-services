// Package main runs the service for indexing fields
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	providerlib "github.com/pokt-foundation/pocket-go/provider"
	indexerlib "github.com/pokt-foundation/pocket-indexer-lib/indexer"
	postgresdriver "github.com/pokt-foundation/pocket-indexer-lib/postgres-driver"
	"github.com/pokt-foundation/pocket-indexer-lib/types"
	"github.com/pokt-foundation/utils-go/environment"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

const (
	concurrentFunctions = 5
)

var (
	errToHeightLowerThanFromHeight          = errors.New("to height is lower than from height")
	errInputHeightIsHigherThanCurrentHeight = errors.New("input height is higher than current height")

	indexingProcesses sync.WaitGroup
	semaphoreLimiter  *semaphore.Weighted

	log = logrus.New()

	clientTimeout    = environment.GetInt64("CLIENT_TIMEOUT", 60000)
	clientRetries    = environment.GetInt64("CLIENT_RETRIES", 3)
	serviceRetries   = environment.GetInt64("SERVICE_RETRIES", 3)
	concurrency      = environment.GetInt64("CONCURRENCY", 100)
	reqInterval      = environment.GetInt64("REQUEST_INTERVAL", 5000)
	connectionString = environment.MustGetString("CONNECTION_STRING")
	mainNode         = environment.MustGetString("MAIN_NODE")
	fallbackNode     = environment.MustGetString("FALLBACK_NODE")
	fromHeight       = int(environment.GetInt64("FROM_HEIGHT", -1))
	toHeight         = int(environment.GetInt64("TO_HEIGHT", -1))
)

func init() {
	// log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&logrus.JSONFormatter{})
}

// indexer interface of needed functions for indexing
type indexer interface {
	IndexBlockTransactions(blockHeight int) error
	IndexBlock(blockHeight int) error
	IndexBlockNodes(blockHeight int) ([]string, error)
	IndexBlockApps(blockHeight int) ([]string, error)
	IndexAccounts(blockHeight int) ([]string, error)
	IndexBlockCalculatedFields(blockHeight int, getTook bool) error
}

// provider interface of needed functions in the provider
type provider interface {
	GetBlock(blockNumber int) (*providerlib.GetBlockOutput, error)
	GetBlockTransactions(options *providerlib.GetBlockTransactionsOptions) (*providerlib.GetBlockTransactionsOutput, error)
	GetAccounts(options *providerlib.GetAccountsOptions) (*providerlib.GetAccountsOutput, error)
	GetNodes(options *providerlib.GetNodesOptions) (*providerlib.GetNodesOutput, error)
	GetApps(options *providerlib.GetAppsOptions) (*providerlib.GetAppsOutput, error)
	GetBlockHeight() (int, error)
	UpdateRequestConfig(retries int, timeout time.Duration)
}

// driver interface of needed functions for the db driver
type driver interface {
	GetMaxHeightInBlocks() (int64, error)
	WriteBlock(block *types.Block) error
	WriteTransactions(txs []*types.Transaction) error
	WriteAccounts(accounts []*types.Account) error
	WriteNodes(nodes []*types.Node) error
	WriteApps(apps []*types.App) error
	GetAccountsQuantity(options *types.GetAccountsQuantityOptions) (int64, error)
	GetAppsQuantity(options *types.GetAppsQuantityOptions) (int64, error)
	GetNodesQuantity(options *types.GetNodesQuantityOptions) (int64, error)
	ReadBlockByHeight(height int) (*types.Block, error)
	WriteBlockCalculatedFields(block *types.Block) error
}

// service struct handler for all necessary fiels for indexing
type service struct {
	indexer          indexer
	fallbackIndexer  indexer
	provider         provider
	fallbackProvider provider
	driver           driver
	hasEnd           bool
	fromHeight       int
	toHeight         int
	retries          int64
	concurrency      int64
	reqInterval      time.Duration
	mainNode         string
	fallbackNode     string
}

func (s *service) logErrorWithFields(message string, height int, err error) {
	fields := logrus.Fields{
		"main_node":     s.mainNode,
		"fallback_node": s.fallbackNode,
		"err":           err.Error(),
	}

	if height > 0 {
		fields["height"] = height
	}

	log.WithFields(fields).Error(fmt.Sprintf("%s with error: %s", message, err.Error()))
}

func (s *service) logInfoWithFields(message, address string, height int) {
	fields := logrus.Fields{
		"main_node":     s.mainNode,
		"fallback_node": s.fallbackNode,
		"height":        height,
	}

	if address != "" {
		fields["address"] = address
	}

	log.WithFields(fields).Info(fmt.Sprintf("%s with height: %d", message, height))
}

func (s *service) start() error {
	for {
		heightsToIndex, err := s.getHeightsToIndex()
		if err != nil {
			return err
		}

		err = s.indexHeights(heightsToIndex)
		if err != nil {
			return err
		}

		if s.hasEnd {
			break
		}

		time.Sleep(s.reqInterval)
	}

	return nil
}

func (s *service) getHeightsToIndex() ([]int, error) {
	maxSavedHeight, err := s.driver.GetMaxHeightInBlocks()
	if err != nil && !errors.Is(err, postgresdriver.ErrNoPreviousHeight) {
		return nil, err
	}

	fromHeight := s.getFromHeight(maxSavedHeight, err)

	currentHeight, err := s.provider.GetBlockHeight()
	if err != nil {
		currentHeight, err = s.fallbackProvider.GetBlockHeight()
		if err != nil {
			return nil, err
		}
	}

	if s.hasEnd && s.toHeight > currentHeight {
		return nil, errInputHeightIsHigherThanCurrentHeight
	}

	toHeight := s.getToHeight(currentHeight)

	var heightsToIndex []int

	for i := fromHeight; i <= toHeight; i++ {
		heightsToIndex = append(heightsToIndex, i)
	}

	return heightsToIndex, nil
}

func (s *service) getFromHeight(maxSavedHeight int64, getMaxHeightErr error) int {
	if s.hasEnd {
		return s.fromHeight
	}

	// This error means nothing was saved in the database and it should start indexing from 1
	if errors.Is(getMaxHeightErr, postgresdriver.ErrNoPreviousHeight) {
		return 1
	}

	return int(maxSavedHeight) + 1
}

func (s *service) getToHeight(currentHeight int) int {
	if s.hasEnd {
		return s.toHeight
	}

	return currentHeight
}

func (s *service) indexHeights(heightsToIndex []int) error {
	if len(heightsToIndex) == 0 {
		return nil
	}

	semaphoreLimiter = semaphore.NewWeighted(s.concurrency)

	for _, height := range heightsToIndex {
		indexingProcesses.Add(concurrentFunctions)

		err := semaphoreLimiter.Acquire(context.Background(), concurrentFunctions)
		if err != nil {
			return err
		}

		go s.indexBlock(height)
		go s.indexAccounts(height)
		go s.indexBlockTransactions(height)
		go s.indexBlockNodes(height)
		go s.indexBlockApps(height)
	}

	indexingProcesses.Wait()

	for _, height := range heightsToIndex {
		indexingProcesses.Add(1)

		err := semaphoreLimiter.Acquire(context.Background(), 1)
		if err != nil {
			return err
		}

		go s.indexBlockCalculatedFields(height, s.getTook(height))
	}

	indexingProcesses.Wait()

	return nil
}

func (s *service) getTook(height int) bool {
	if !s.hasEnd {
		return true
	}

	return height != s.fromHeight
}

func releaseProcess() {
	indexingProcesses.Done()
	semaphoreLimiter.Release(1)
}

func (s *service) indexBlock(height int) {
	defer releaseProcess()

	err := s.indexBlockWithRetries(height, s.indexer)
	if err != nil {
		s.logErrorWithFields("Index block with main node failed", height, err)

		err = s.indexBlockWithRetries(height, s.fallbackIndexer)
		if err != nil {
			s.logErrorWithFields("Index block with fallback node failed", height, err)

			return
		}
	}

	s.logInfoWithFields("Block indexed successfully", "", height)
}

func (s *service) indexBlockWithRetries(height int, indexer indexer) error {
	var retry int64
	var err error

	for {
		err = indexer.IndexBlock(height)
		if err == nil {
			break
		}

		retry++

		if retry == s.retries {
			break
		}
	}

	return err
}

func (s *service) indexBlockCalculatedFields(height int, getTook bool) {
	defer releaseProcess()

	err := s.indexBlockCalculatedFieldsWithRetries(height, getTook, s.indexer)
	if err != nil {
		s.logErrorWithFields("Index block calculated fields failed", height, err)

		return
	}

	s.logInfoWithFields("Block calculated fields indexed successfully", "", height)
}

func (s *service) indexBlockCalculatedFieldsWithRetries(height int, getTook bool, indexer indexer) error {
	var retry int64
	var err error

	for {
		err = indexer.IndexBlockCalculatedFields(height, getTook)
		if err == nil {
			break
		}

		retry++

		if retry == s.retries {
			break
		}
	}

	return err
}

func (s *service) indexAccounts(height int) {
	defer releaseProcess()

	err := s.indexAccountsWithRetries(height, s.indexer)
	if err != nil {
		s.logErrorWithFields("Index accounts with main node failed", height, err)

		err = s.indexAccountsWithRetries(height, s.fallbackIndexer)
		if err != nil {
			s.logErrorWithFields("Index accounts with fallback node failed", height, err)

			return
		}
	}

	s.logInfoWithFields("Block accounts indexed successfully", "", height)
}

func (s *service) indexAccountsWithRetries(height int, indexer indexer) error {
	var retry int64
	var err error

	for {
		_, err = indexer.IndexAccounts(height)
		if err == nil {
			break
		}

		retry++

		if retry == s.retries {
			break
		}
	}

	return err
}

func (s *service) indexBlockTransactions(height int) {
	defer releaseProcess()

	err := s.indexBlockTransactionsWithRetries(height, s.indexer)
	if err != nil {
		s.logErrorWithFields("Index block transactions with main node failed", height, err)

		err = s.indexBlockTransactionsWithRetries(height, s.fallbackIndexer)
		if err != nil {
			s.logErrorWithFields("Index block transactions with fallback node failed", height, err)

			return
		}
	}

	s.logInfoWithFields("Block transactions indexed successfully", "", height)
}

func (s *service) indexBlockTransactionsWithRetries(height int, indexer indexer) error {
	var retry int64
	var err error

	for {
		err = indexer.IndexBlockTransactions(height)
		if err == nil {
			break
		}

		retry++

		if retry == s.retries {
			break
		}
	}

	return err
}

func (s *service) indexBlockNodes(height int) {
	defer releaseProcess()

	err := s.indexBlockNodesWithRetries(height, s.indexer)
	if err != nil {
		s.logErrorWithFields("Index nodes with main node failed", height, err)

		err = s.indexBlockNodesWithRetries(height, s.fallbackIndexer)
		if err != nil {
			s.logErrorWithFields("Index nodes with fallback node failed", height, err)

			return
		}
	}

	s.logInfoWithFields("Block nodes indexed successfully", "", height)
}

func (s *service) indexBlockNodesWithRetries(height int, indexer indexer) error {
	var retry int64
	var err error

	for {
		_, err = indexer.IndexBlockNodes(height)
		if err == nil {
			break
		}

		retry++

		if retry == s.retries {
			break
		}
	}

	return err
}

func (s *service) indexBlockApps(height int) {
	defer releaseProcess()

	err := s.indexBlockAppsWithRetries(height, s.indexer)
	if err != nil {
		s.logErrorWithFields("Index apps with main node failed", height, err)

		err = s.indexBlockAppsWithRetries(height, s.fallbackIndexer)
		if err != nil {
			s.logErrorWithFields("Index apps with fallback node failed", height, err)

			return
		}
	}

	s.logInfoWithFields("Block apps indexed successfully", "", height)
}

func (s *service) indexBlockAppsWithRetries(height int, indexer indexer) error {
	var retry int64
	var err error

	for {
		_, err = indexer.IndexBlockApps(height)
		if err == nil {
			break
		}

		retry++

		if retry == s.retries {
			break
		}
	}

	return err
}

func getFallbacks(fallbackNode string, driver driver) (provider, indexer) {
	if fallbackNode == "" {
		return nil, nil
	}

	fallbackProvider := providerlib.NewProvider(fallbackNode, nil)

	fallbackProvider.UpdateRequestConfig(int(clientRetries), time.Duration(clientTimeout)*time.Millisecond)

	fallbackIndexer := indexerlib.NewIndexer(fallbackProvider, driver)

	return fallbackProvider, fallbackIndexer
}

func (s *service) setOptionalParams(fromHeight, toHeight int) error {
	if fromHeight > 0 && toHeight > 0 {
		if toHeight < fromHeight {
			return errToHeightLowerThanFromHeight
		}

		s.hasEnd = true
		s.fromHeight = fromHeight
		s.toHeight = toHeight
	}

	return nil
}

func setupService() (*service, error) {
	mainProvider := providerlib.NewProvider(mainNode, nil)

	mainProvider.UpdateRequestConfig(int(clientRetries), time.Duration(clientTimeout)*time.Millisecond)

	driver, err := postgresdriver.NewPostgresDriverFromConnectionString(connectionString)
	if err != nil {
		return nil, err
	}

	mainIndexer := indexerlib.NewIndexer(mainProvider, driver)

	fallbackProvider, fallbackIndexer := getFallbacks(fallbackNode, driver)

	service := &service{
		indexer:          mainIndexer,
		fallbackIndexer:  fallbackIndexer,
		provider:         mainProvider,
		fallbackProvider: fallbackProvider,
		driver:           driver,
		retries:          serviceRetries,
		concurrency:      concurrency,
		reqInterval:      time.Duration(reqInterval) * time.Millisecond,
		mainNode:         mainNode,
		fallbackNode:     fallbackNode,
	}

	err = service.setOptionalParams(fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func main() {
	service, err := setupService()
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Setup service failed with error: %s", err.Error()))
	}

	err = service.start()
	if err != nil {
		service.logErrorWithFields("Start service failed", -1, err)
	}

	log.Info("Execution finished successfully")
}
