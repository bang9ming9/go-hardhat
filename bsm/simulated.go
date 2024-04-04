package bsm

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/bang9ming9/go-hardhat/bsm/utils"
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

func ChainID() *big.Int {
	return params.AllDevChainProtocolChanges.ChainID
}

type Backend struct {
	simulated.Backend
	simulated.Client
	Owner *bind.TransactOpts
}

func NewBacked(t *testing.T) *Backend {
	owner := GetOwner(t)

	backend := simulated.NewBackend(
		core.GenesisAlloc{
			owner.From: core.GenesisAccount{Balance: utils.ToWei(common.Big256)},
		},
		simulated.WithBlockGasLimit(params.MaxGasLimit),
		func(nodeConf *node.Config, ethConf *ethconfig.Config) {
			ethConf.Genesis.Coinbase = owner.From
			ethConf.Genesis.BaseFee = common.Big0
			// @dev ethconfig.Defaults.Miner.GasPrice = common.Big0 을 통해서 완전한 공짜 블록체인이 완성되었다.
			// @dev ethconfig.Defaults.Miner.GasPrice 를 또 어디서 이용하는지 확인해봐야 겠다.
			// @dev ethereum@1.13.11 에서는 해당 내용은 필요 없었다.
			ethconfig.Defaults.Miner.GasPrice = common.Big0
			ethConf.Miner.GasPrice = common.Big0

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
	return gas, ToRevert(err)
}

func (ec *Backend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return ToRevert(ec.Client.SendTransaction(ctx, tx))
}
