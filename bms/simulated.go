package bms

import (
	"context"
	"errors"
	"math/big"
	"testing"

	bmsutils "github.com/bang9ming9/go-hardhat/bms/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
)

var ChainID *big.Int = params.AllDevChainProtocolChanges.ChainID

type Backend struct {
	simulated.Backend
	simulated.Client
	Owner *bind.TransactOpts
}

func NewBacked(t *testing.T) *Backend {
	minerGasPrice := new(big.Int).SetBytes(ethconfig.Defaults.Miner.GasPrice.Bytes())
	ethconfig.Defaults.Miner.GasPrice.SetBytes([]byte{})
	defer ethconfig.Defaults.Miner.GasPrice.SetBytes(minerGasPrice.Bytes())

	owner := GetOwner(t)
	ethconfig.Defaults.Miner.GasPrice = common.Big0
	backend := simulated.NewBackend(
		core.GenesisAlloc{
			owner.From: core.GenesisAccount{Balance: bmsutils.ToWei(common.Big256)},
		},
		simulated.WithBlockGasLimit(params.MaxGasLimit),
		func(nodeConf *node.Config, ethConf *ethconfig.Config) {
			ethConf.Genesis.Coinbase = owner.From
			ethConf.Genesis.BaseFee = common.Big0
			ethConf.TxPool.PriceLimit = 0
		},
	)

	return &Backend{
		Backend: *backend,
		Client:  backend.Client(),
		Owner:   owner,
	}
}

func (ec *Backend) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return common.Big0, nil
}

func (ec *Backend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return common.Big0, nil
}

func (ec *Backend) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	receipt, err := ec.Client.TransactionReceipt(ctx, txHash)
	if err != nil && (errors.Is(err, ethereum.NotFound) || err.Error() == "transaction indexing is in progress") {
		ec.Commit()
	}
	return receipt, err
}

func (ec *Backend) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	gas, err := ec.Client.EstimateGas(ctx, call)
	return gas, bmsutils.ToRevert(err)
}

func (ec *Backend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return bmsutils.ToRevert(ec.Client.SendTransaction(ctx, tx))
}
