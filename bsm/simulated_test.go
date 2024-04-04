package bsm_test

import (
	"context"
	"testing"

	"github.com/bang9ming9/go-hardhat/bsm"
	"github.com/stretchr/testify/require"
)

func TestNewBackend(t *testing.T) {
	backend := bsm.NewBacked(t)
	ctx := context.Background()
	balance, err := backend.BalanceAt(ctx, backend.Owner.From, nil)
	require.NoError(t, err)
	t.Log(backend.Owner.From, "balance", balance)

	eoa := bsm.GetEOA(t)
	balance, err = backend.BalanceAt(ctx, eoa.From, nil)
	require.NoError(t, err)
	t.Log(eoa.From, "balance", balance)

}
