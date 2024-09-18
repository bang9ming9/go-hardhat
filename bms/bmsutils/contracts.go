package bmsutils

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Contract[T any] struct {
	address common.Address
	funcs   *T
	abi     *abi.ABI
}

func DeployContract[T any](address common.Address, tx *types.Transaction, contract *T, err error) (*Contract[T], *types.Transaction, error) {
	if err != nil {
		return nil, nil, err
	}
	return &Contract[T]{address: address, funcs: contract}, tx, nil
}

func NewContract[T any](contract *T, err error) (*Contract[T], error) {
	if err != nil {
		return nil, err
	}
	funcs := new(T)
	return &Contract[T]{address: common.Address{0}, funcs: funcs}, nil
}

func (contract *Contract[T]) SetAddress(address common.Address) *Contract[T] {
	if contract.address == (common.Address{0}) {
		contract.address = address
	}
	return contract
}

func (contract *Contract[T]) SetABI(abi *abi.ABI) *Contract[T] {
	if contract.abi == nil {
		EnrollErrors(abi)
		contract.abi = abi
	}
	return contract
}

func (contract *Contract[T]) SetABIWithError(abi *abi.ABI, err error) (*Contract[T], error) {
	if err != nil {
		return nil, err
	}
	return contract.SetABI(abi), nil
}

func (contract *Contract[T]) Address() common.Address {
	return contract.address
}

func (contract *Contract[T]) Funcs() *T {
	return contract.funcs
}

func (contract *Contract[T]) ABI() *abi.ABI {
	return contract.abi
}
