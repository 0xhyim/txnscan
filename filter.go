package txnscan

import (
	"encoding/hex"
	"math/big"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type TransactionFilters []TransactionFilter

func (f TransactionFilters) evaluate(t *types.Transaction) bool {
	if len(f) == 0 {
		return true
	}

	for _, filter := range f {
		if filter.evaluate(t) {
			return true
		}
	}

	return false
}

type TransactionFilter interface {
	evaluate(*types.Transaction) bool
	And(TransactionFilter) TransactionFilter
}

type transactionFilter struct {
	eval func(*types.Transaction) bool
}

func newTransactionFilter(eval func(*types.Transaction) bool) *transactionFilter {
	return &transactionFilter{
		eval: eval,
	}
}

func (f *transactionFilter) evaluate(t *types.Transaction) bool {
	return f.eval(t)
}

func (f *transactionFilter) And(filter TransactionFilter) TransactionFilter {
	originalEval := f.eval
	f.eval = func(t *types.Transaction) bool {
		return originalEval(t) && filter.evaluate(t)
	}
	return f
}

func WithCustomFilter(eval func(t *types.Transaction) bool) TransactionFilter {
	return newTransactionFilter(eval)
}

func WithToAddress(addr string) TransactionFilter {
	return newTransactionFilter(func(t *types.Transaction) bool {
		if t.To() == nil {
			return false
		}
		return t.To().Hex() == common.HexToAddress(addr).Hex()
	})
}

func WithFromAddress(addr string) TransactionFilter {
	return newTransactionFilter(func(t *types.Transaction) bool {
		from, err := types.Sender(types.LatestSignerForChainID(t.ChainId()), t)
		if err != nil {
			return false
		}
		return from.Hex() == common.HexToAddress(addr).Hex()
	})
}

func WithValueGreaterThan(n *big.Int) TransactionFilter {
	return newTransactionFilter(func(t *types.Transaction) bool {
		return t.Value().Cmp(n) == 1
	})
}

func WithValueLessThan(n *big.Int) TransactionFilter {
	return newTransactionFilter(func(t *types.Transaction) bool {
		return t.Value().Cmp(n) == -1
	})
}

func WithCostGreaterThan(n *big.Int) TransactionFilter {
	return newTransactionFilter(func(t *types.Transaction) bool {
		return t.Cost().Cmp(n) == 1
	})
}

func WithCostLessThan(n *big.Int) TransactionFilter {
	return newTransactionFilter(func(t *types.Transaction) bool {
		return t.Cost().Cmp(n) == -1
	})
}

func WithFunctionSignature(signature string) TransactionFilter {
	return newTransactionFilter(func(t *types.Transaction) bool {
		signature = strings.TrimPrefix(signature, "0x")
		for chunk := range slices.Chunk(t.Data(), 4) {
			return hex.EncodeToString(chunk) == signature
		}
		return false
	})
}
