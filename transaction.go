package txnscan

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Transaction struct {
	types.Transaction
	from common.Address
}

func (t *Transaction) From() *common.Address {
	return &t.from
}
