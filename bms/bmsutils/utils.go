package bmsutils

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

var (
	zero = decimal.NewFromInt(0)
	wei  = decimal.NewFromBigInt(common.Big1, 18)
)

func ToWei(x interface{}) *big.Int {
	v, err := decimal.NewFromString(fmt.Sprint(x))
	if err != nil {
		v = zero
	}
	return v.Mul(wei).BigInt()
}

func ToEther(x interface{}) decimal.Decimal {
	v, err := decimal.NewFromString(fmt.Sprint(x))
	if err != nil {
		v = zero
	}
	return v.Div(wei)
}
