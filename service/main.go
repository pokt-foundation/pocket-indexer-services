package main

import (
	"context"
	"errors"
	"os"
	"strconv"
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

	indexingProcesses  sync.WaitGroup
	concurrencyLimiter *semaphore.Weighted

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
	indexer     indexer
	provider    provider
	driver      driver
	isCustom    bool
	fromHeight  int
	toHeight    int
	retries     int64
	concurrency int64
	reqInterval time.Duration
}

func newService(retries, concurrency int64, reqInterval time.Duration, provider provider, driver driver, indexer indexer, options *indexOptions) (*service, error) {
	service := &service{
		indexer:     indexer,
		provider:    provider,
		driver:      driver,
		retries:     retries,
		concurrency: concurrency,
		reqInterval: reqInterval,
	}

	if options != nil {
		if options.toHeight < options.fromHeight {
			return nil, errToHeightLowerThanFromHeight
		}

		service.isCustom = true
		service.fromHeight = options.fromHeight
		service.toHeight = options.toHeight
	}

	return service, nil
}

func (s *service) start() error {
	for {
		heightsToIndex, err := s.getHeightsToIndex()
		if err != nil {
			return err
		}

		s.indexHeights(heightsToIndex)

		// Just custom indexer should stop, the other ones keep looking if a new blocks comes up to index it
		if s.isCustom {
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
		return nil, err
	}

	if s.isCustom && s.toHeight > currentHeight {
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
	if s.isCustom {
		return s.fromHeight
	}

	// This error means nothing was saved in the database and it should start indexing from 0
	if errors.Is(getMaxHeightErr, postgresdriver.ErrNoPreviousHeight) {
		return 0
	}

	return int(maxSavedHeight) + 1
}

func (s *service) getToHeight(currentHeight int) int {
	if s.isCustom {
		return s.toHeight
	}

	return currentHeight
}

func (s *service) indexHeights(heightsToIndex []int) error {
	if len(heightsToIndex) == 0 {
		return nil
	}

	concurrencyLimiter = semaphore.NewWeighted(s.concurrency)

	for _, height := range heightsToIndex {
		indexingProcesses.Add(2)

		err := concurrencyLimiter.Acquire(context.Background(), 2)
		if err != nil {
			return err
		}

		go s.indexBlockWithRetries(height)
		go s.indexBlockTransactionsWithRetries(height)
	}

	indexingProcesses.Wait()

	return nil
}

func (s *service) indexBlockWithRetries(height int) {
	defer indexingProcesses.Done()
	defer concurrencyLimiter.Release(1)

	// Height 0 does not a have a block
	// Core always returns error with it
	if height == 0 {
		return
	}

	var retry int64

	for {
		err := s.indexer.IndexBlock(height)
		if err == nil {
			break
		}

		retry++

		if retry == s.retries {
			break
		}
	}
}

func (s *service) indexBlockTransactionsWithRetries(height int) {
	defer indexingProcesses.Done()
	defer concurrencyLimiter.Release(1)

	var retry int64

	for {
		err := s.indexer.IndexBlockTransactions(height)
		if err == nil {
			break
		}

		retry++

		if retry == s.retries {
			break
		}
	}
}

func setupService() (*service, error) {
	reqProvider := providerlib.NewProvider(os.Args[1], nil)

	reqProvider.UpdateRequestConfig(int(clientRetries), time.Duration(clientTimeout)*time.Millisecond)

	driver, err := postgresdriver.NewPostgresDriverFromConnectionString(connectionString)
	if err != nil {
		return nil, err
	}

	indexer := indexerlib.NewIndexer(reqProvider, driver)

	options := &indexOptions{}

	if len(os.Args) > 2 {
		fromHeight, err := strconv.Atoi(os.Args[2])
		if err != nil {
			return nil, err
		}

		toHeight, err := strconv.Atoi(os.Args[3])
		if err != nil {
			return nil, err
		}

		options = &indexOptions{
			fromHeight: fromHeight,
			toHeight:   toHeight,
		}
	} else {
		options = nil
	}

	return newService(serviceRetries, concurrency, time.Duration(reqInterval)*time.Millisecond,
		reqProvider, driver, indexer, options)
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
