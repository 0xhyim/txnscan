package txnscan

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var defaultTimeout = 15 * time.Second

type scanner struct {
	client  *ethclient.Client
	filters []TransactionFilter
}

func NewScanner(ctx context.Context, networkUrl string, filters ...TransactionFilter) (*scanner, error) {
	client, err := ethclient.DialContext(ctx, networkUrl)
	return &scanner{client, filters}, err
}

func (s *scanner) SubscribeNewTransactions(ctx context.Context, transactions chan<- Transaction) (ethereum.Subscription, error) {
	headers := make(chan *types.Header)
	sub, err := s.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return nil, err
	}
	go s.startScanning(sub, headers, transactions)
	return sub, err
}

func (s *scanner) startScanning(sub ethereum.Subscription, headers chan *types.Header, transactions chan<- Transaction) {
	defer close(headers)
	for {
		select {
		case <-sub.Err():
			return
		case header := <-headers:
			txns := s.transactionsByHeader(header)
			sendTransactions(s.filter(txns), transactions)
		}
	}
}

func (s *scanner) transactionsByHeader(header *types.Header) (txns []Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	block, err := s.client.BlockByHash(ctx, header.Hash())
	if err != nil {
		return []Transaction{}
	}
	for _, txn := range block.Transactions() {
		message, err := txn.AsMessage(types.LatestSignerForChainID(txn.ChainId()), block.BaseFee())
		if err != nil {
			continue
		}
		txns = append(txns, Transaction{Transaction: *txn, from: message.From()})
	}
	return
}

func (s *scanner) filter(txns []Transaction) (ret []Transaction) {
	if len(s.filters) == 0 {
		return txns
	}
	for _, txn := range txns {
		for _, filter := range s.filters {
			if filter.evaluate(&txn) {
				ret = append(ret, txn)
			}
		}
	}
	return ret
}

func sendTransactions(txns []Transaction, ch chan<- Transaction) {
	for _, txn := range txns {
		ch <- txn
	}
}
