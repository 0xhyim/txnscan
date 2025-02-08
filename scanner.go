package txnscan

import (
	"context"
	"iter"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type scanner struct {
	networkUrl string
}

func NewScanner(networkUrl string) *scanner {
	return &scanner{networkUrl}
}

func (s *scanner) SubscribeNewTransactions(ctx context.Context, txCh chan<- *types.Transaction, filters ...TransactionFilter) (ethereum.Subscription, error) {
	client, err := ethclient.DialContext(ctx, s.networkUrl)
	if err != nil {
		return nil, err
	}

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return nil, err
	}

	go s.startScan(client, sub, headers, txCh, filters...)

	return sub, err
}

func (s *scanner) startScan(client *ethclient.Client, sub ethereum.Subscription, headers chan *types.Header, txCh chan<- *types.Transaction, filters ...TransactionFilter) {
	defer close(headers)
	for {
		select {
		case <-sub.Err():
			return
		case header := <-headers:
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			block, err := client.BlockByHash(ctx, header.Hash())
			if err != nil {
				continue
			}

			for tx := range filteredTransactions(filters, block.Transactions()) {
				txCh <- tx
			}
		}
	}
}

func filteredTransactions(filters TransactionFilters, txs types.Transactions) iter.Seq[*types.Transaction] {
	return func(yield func(*types.Transaction) bool) {
		for _, tx := range txs {
			if !filters.evaluate(tx) {
				continue
			}
			if !yield(tx) {
				return
			}
		}
	}
}
