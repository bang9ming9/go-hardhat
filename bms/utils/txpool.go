package bmsutils

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

type TxPool struct {
	client bind.DeployBackend
	txs    []*types.Transaction
}

func NewTxPool(client bind.DeployBackend) *TxPool {
	return &TxPool{client: client, txs: make([]*types.Transaction, 0)}
}

func (txs *TxPool) Exec(tx *types.Transaction, err error) error {
	if tx != nil {
		txs.txs = append(txs.txs, tx)
	}
	return err
}

func (txs *TxPool) Append(tx *types.Transaction) *types.Transaction {
	if tx != nil {
		txs.txs = append(txs.txs, tx)
	}
	return tx
}

func (txs *TxPool) AllReceiptStatusSuccessful(ctx context.Context) error {
	receipts, err := txs.WaitMined(ctx)
	if err != nil {
		return err
	}
	notsuccesses := make([]common.Hash, 0)
	for _, receipt := range receipts {
		if receipt.Status != types.ReceiptStatusSuccessful {
			notsuccesses = append(notsuccesses, receipt.TxHash)
		}
	}
	if len(notsuccesses) == 0 {
		return nil
	}
	return fmt.Errorf("not successes: %v", notsuccesses)
}

func (txs *TxPool) WaitMined(ctx context.Context) ([]*types.Receipt, error) {
	receipts := make([]*types.Receipt, 0)
	length := len(txs.txs)
	for i := 0; i < length; i++ {
		ctx, cancel := context.WithTimeout(ensureContext(ctx), 10e9)
		receipt, err := bind.WaitMined(ctx, txs.client, txs.txs[i])
		cancel()

		if err != nil {
			txs.txs = txs.txs[i:]
			return receipts, errors.Wrap(err, "bind.WaitMined")
		}
		receipts = append(receipts, receipt)
	}
	txs.txs = make([]*types.Transaction, 0)
	return receipts, nil
}

func (txs *TxPool) Clear() error {
	length := len(txs.txs)
	if length == 0 {
		return nil
	}
	return fmt.Errorf("ignore tx counts : %d", length)
}
