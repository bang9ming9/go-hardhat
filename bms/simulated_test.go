package bms_test

import (
	"context"
	"testing"

	"github.com/bang9ming9/go-hardhat/bms"
	bmsutils "github.com/bang9ming9/go-hardhat/bms/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestNewBackend(t *testing.T) {
	backend := bms.NewBacked(t)
	ctx := context.Background()
	balance, err := backend.BalanceAt(ctx, backend.Owner.From, nil)
	require.NoError(t, err)
	t.Log(backend.Owner.From, "balance", balance)

	eoa := bms.GetEOA(t)
	balance, err = backend.BalanceAt(ctx, eoa.From, nil)
	require.NoError(t, err)
	t.Log(eoa.From, "balance", balance)

	txpool := bmsutils.NewTxPool(backend)

	backend.Owner.Value = bmsutils.ToWei(1)
	require.NoError(t, txpool.Exec(bmsutils.SendDynamicTx(backend, backend.Owner, &eoa.From, []byte{})))
	backend.Owner.Value = common.Big0

	eoa.Value = bmsutils.ToWei(2)
	require.Error(t, txpool.Exec(bmsutils.SendDynamicTx(backend, eoa, &backend.Owner.From, []byte{})))
	eoa.Value = common.Big0

	receipts, err := txpool.WaitMined(ctx)
	require.NoError(t, err)
	for _, receipt := range receipts {
		require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	}

	balance, err = backend.BalanceAt(ctx, backend.Owner.From, nil)
	require.NoError(t, err)
	t.Log(backend.Owner.From, "balance", balance)

	balance, err = backend.BalanceAt(ctx, eoa.From, nil)
	require.NoError(t, err)
	t.Log(eoa.From, "balance", balance)
}
