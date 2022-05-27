package main

import (
	"context"
	"errors"
	"flag"
	"sync"
	"time"

	providerlib "github.com/pokt-foundation/pocket-go/provider"
	indexerlib "github.com/pokt-foundation/pocket-indexer-lib"
	postgresdriver "github.com/pokt-foundation/pocket-indexer-lib/postgres-driver"
	"github.com/pokt-foundation/pocket-indexer-services/pkg/environment"
	"golang.org/x/sync/semaphore"
)

var (
	errToHeightLowerThanFromHeight          = errors.New("to height is lower than from height")
	errInputHeightIsHigherThanCurrentHeight = errors.New("input height is higher than current height")

	indexingProcesses sync.WaitGroup
	semaphoreLimiter  *semaphore.Weighted

	clientTimeout    = environment.GetInt64("CLIENT_TIMEOUT", 60000)
	clientRetries    = environment.GetInt64("CLIENT_RETRIES", 3)
	serviceRetries   = environment.GetInt64("SERVICE_RETRIES", 3)
	concurrency      = environment.GetInt64("CONCURRENCY", 100)
	reqInterval      = environment.GetInt64("REQUEST_INTERVAL", 5000)
	connectionString = environment.GetString("CONNECTION_STRING", "")
)

// indexer interface of needed functions for indexing
type indexer interface {
	IndexBlockTransactions(blockHeight int) error
	IndexBlock(blockHeight int) error
}

// provider interface of needed functions in the provider
type provider interface {
	GetBlock(blockNumber int) (*providerlib.GetBlockOutput, error)
	GetBlockTransactions(blockHeight int, options *providerlib.GetBlockTransactionsOptions) (*providerlib.GetBlockTransactionsOutput, error)
	GetBlockHeight() (int, error)
	UpdateRequestConfig(retries int, timeout time.Duration)
}

// driver interface of needed functions for the db driver
type driver interface {
	GetMaxHeightInBlocks() (int64, error)
	WriteBlock(block *indexerlib.Block) error
	WriteTransactions(txs []*indexerlib.Transaction) error
}

// indexOptions optional parameters for the custom range in indexing
type indexOptions struct {
	fromHeight int
	toHeight   int
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
}

func (s *service) start() error {
	for {
		heightsToIndex, err := s.getHeightsToIndex()
		if err != nil {
			return err
		}

		s.indexHeights(heightsToIndex)

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

	// This error means nothing was saved in the database and it should start indexing from 0
	if errors.Is(getMaxHeightErr, postgresdriver.ErrNoPreviousHeight) {
		return 0
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
		indexingProcesses.Add(2)

		err := semaphoreLimiter.Acquire(context.Background(), 2)
		if err != nil {
			return err
		}

		go s.indexBlock(height)
		go s.indexBlockTransactions(height)
	}

	indexingProcesses.Wait()

	return nil
}

func releaseProcess() {
	indexingProcesses.Done()
	semaphoreLimiter.Release(1)
}

func (s *service) indexBlock(height int) {
	defer releaseProcess()

	// Height 0 does not a have a block
	// Core always returns error with it
	if height == 0 {
		return
	}

	err := s.indexBlockWithRetries(height, s.indexer)
	if err != nil {
		s.indexBlockWithRetries(height, s.fallbackIndexer)
	}
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

func (s *service) indexBlockTransactions(height int) {
	defer releaseProcess()

	err := s.indexBlockTransactionsWithRetries(height, s.indexer)
	if err != nil {
		s.indexBlockTransactionsWithRetries(height, s.fallbackIndexer)
	}
}

func (s *service) indexBlockTransactionsWithRetries(height int, indexer indexer) error {
	var retry int64
	var err error

	for {
		err := indexer.IndexBlockTransactions(height)
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

func parseParams() (string, string, int, int) {
	mainNode := flag.String("node", "", "Main node URL to index")
	fallbackNode := flag.String("fallback", "", "Fallback node URL to index in case main one fails")
	fromHeight := flag.Int("from", -1, "Starting height to index, optional param")
	toHeight := flag.Int("to", -1, "Final height to index, optional param")

	flag.Parse()

	return *mainNode, *fallbackNode, *fromHeight, *toHeight
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

func getOptions(fromHeight, toHeight int) *indexOptions {
	if fromHeight >= 0 && toHeight >= 0 {
		return &indexOptions{
			fromHeight: fromHeight,
			toHeight:   toHeight,
		}
	}

	return nil
}

func (s *service) setOptionalParams(fromHeight, toHeight int) error {
	if fromHeight >= 0 && toHeight >= 0 {
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
	mainNode, fallbackNode, fromHeight, toHeight := parseParams()

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
		return
	}

	err = service.start()
	if err != nil {
		return
	}
}
