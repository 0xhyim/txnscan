package txnscan

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type TransactionFilter interface {
	evaluate(*Transaction) bool
	And(TransactionFilter) TransactionFilter
}

type transactionFilter struct {
	eval func(*Transaction) bool
}

func newAddressFilter(eval func(*Transaction) bool) *transactionFilter {
	return &transactionFilter{
		eval: eval,
	}
}

func (f *transactionFilter) evaluate(t *Transaction) bool {
	return f.eval(t)
}

func (f *transactionFilter) And(filter TransactionFilter) TransactionFilter {
	originalEval := f.eval
	f.eval = func(t *Transaction) bool {
		return originalEval(t) && filter.evaluate(t)
	}
	return f
}

func WithCustomFilter(eval func(t *Transaction) bool) TransactionFilter {
	return newAddressFilter(eval)
}

func WithToAddress(addr string) TransactionFilter {
	return newAddressFilter(func(t *Transaction) bool {
		return t.To().Hex() == common.HexToAddress(addr).Hex()
	})
}

func WithFromAddress(addr string) TransactionFilter {
	return newAddressFilter(func(t *Transaction) bool {
		return t.From().Hex() == common.HexToAddress(addr).Hex()
	})
}

func WithValueGreaterThan(n *big.Int) TransactionFilter {
	return newAddressFilter(func(t *Transaction) bool {
		return t.Value().Cmp(n) == 1
	})
}

func WithValueLessThan(n *big.Int) TransactionFilter {
	return newAddressFilter(func(t *Transaction) bool {
		return t.Value().Cmp(n) == -1
	})
}

func WithCostGreaterThan(n *big.Int) TransactionFilter {
	return newAddressFilter(func(t *Transaction) bool {
		return t.Cost().Cmp(n) == 1
	})
}

func WithCostLessThan(n *big.Int) TransactionFilter {
	return newAddressFilter(func(t *Transaction) bool {
		return t.Cost().Cmp(n) == -1
	})
}

func WithFunctionSignature(signature string) TransactionFilter {
	return newAddressFilter(func(t *Transaction) bool {
		signature = strings.TrimPrefix(signature, "0x")
		return hex.EncodeToString(parseFuncSignature(t)) == signature
	})
}

func parseFuncSignature(txn *Transaction) []byte {
	if len(txn.Data()) < 4 {
		return txn.Data()
	}
	return txn.Data()[:4]
}
