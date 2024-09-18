package bmsutils

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Sig [4]byte

var (
	errorABIs map[Sig]abi.Error = make(map[Sig]abi.Error)
)

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

type RevertError struct {
	err  abi.Error
	args interface{}
}

func (err *RevertError) Error() string {
	return fmt.Sprintf("%s%v", err.err.Name, err.args)
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
		return input
	}

	return &RevertError{aBI, args}
}
