package bmsutils

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Backend interface {
	bind.ContractBackend
	bind.DeployBackend
}

func CreateDynamicTx(backend Backend, opts *bind.TransactOpts, to *common.Address, input []byte) (*types.Transaction, error) {
	// Normalize value
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	// Estimate TipCap
	gasTipCap := opts.GasTipCap
	if gasTipCap == nil {
		tip, err := backend.SuggestGasTipCap(ensureContext(opts.Context))
		if err != nil {
			return nil, err
		}
		gasTipCap = tip
	}
	// Estimate FeeCap
	gasFeeCap := opts.GasFeeCap
	if gasFeeCap == nil {
		gasFeeCap = gasTipCap
	}
	if gasFeeCap.Cmp(gasTipCap) < 0 {
		return nil, fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", gasFeeCap, gasTipCap)
	}
	// Estimate GasLimit
	gasLimit := opts.GasLimit
	if opts.GasLimit == 0 {
		var err error
		gasLimit, err = estimateGasLimit(backend, opts, to, input, nil, gasTipCap, gasFeeCap, value)
		if err != nil {
			return nil, err
		}
	}
	// create the transaction
	nonce, err := getNonce(backend, opts)
	if err != nil {
		return nil, err
	}
	baseTx := &types.DynamicFeeTx{
		To:        to,
		Nonce:     nonce,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Gas:       gasLimit,
		Value:     value,
		Data:      input,
	}

	return opts.Signer(opts.From, types.NewTx(baseTx))
}

func SendDynamicTx(backend Backend, opts *bind.TransactOpts, to *common.Address, input []byte) (*types.Transaction, error) {
	if tx, err := CreateDynamicTx(backend, opts, to, input); err != nil {
		return nil, err
	} else {
		return tx, backend.SendTransaction(ensureContext(opts.Context), tx)
	}
}

// ensureContext is a helper method to ensure a context is not nil, even if the
// user specified it as such.
func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func estimateGasLimit(backend Backend, opts *bind.TransactOpts, to *common.Address, input []byte, gasPrice, gasTipCap, gasFeeCap, value *big.Int) (uint64, error) {
	msg := ethereum.CallMsg{
		From:      opts.From,
		To:        to,
		GasPrice:  gasPrice,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Value:     value,
		Data:      input,
	}
	return backend.EstimateGas(ensureContext(opts.Context), msg)
}

func getNonce(backend Backend, opts *bind.TransactOpts) (uint64, error) {
	if opts.Nonce == nil {
		return backend.PendingNonceAt(ensureContext(opts.Context), opts.From)
	} else {
		return opts.Nonce.Uint64(), nil
	}
}
