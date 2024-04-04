package bsm

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

type Sig [4]byte

var (
	errorABIs map[Sig]abi.Error
)

func ClearErrors() {
	errorABIs = make(map[Sig]abi.Error)
}

func EnrollErrors(aBIs ...*abi.ABI) {
	for _, aBI := range aBIs {
		for _, err := range aBI.Errors {
			sig := Sig(err.ID[:4])
			if _, ok := errorABIs[sig]; !ok {
				errorABIs[sig] = err
			}
		}
	}
}

type revertError interface {
	ErrorData() interface{}
}

func ToRevert(input error) error {
	if input == nil {
		return nil
	}

	revert, ok := input.(revertError)
	if !ok {
		return input
	}

	hexBytes, ok := revert.ErrorData().(string)
	if !ok {
		return input
	}

	data, err := hexutil.Decode(hexBytes)
	if err != nil || len(data) < 4 {
		return input
	}

	aBI, ok := errorABIs[Sig(data[:4])]
	if !ok {
		return input
	}

	args, err := aBI.Unpack(data)
	if err != nil {
		return errors.Wrap(errors.Wrap(err, "aBI.Unpack"), input.Error())
	}

	return fmt.Errorf("%s: %v", aBI.Sig, args)
}
