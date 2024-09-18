package bms

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/stretchr/testify/require"
)

var (
	mnemonicFile string = filepath.Join(os.Getenv("HOME"), ".bms-mnemonic")
	wallet       *hdwallet.Wallet
	eoaTCount    uint32 = 0
	eoaCount     uint32 = 0
)

func init() {
	if _, err := os.Stat(mnemonicFile); os.IsNotExist(err) {
		mnemonic, err := hdwallet.NewMnemonic(192)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(mnemonicFile, []byte(mnemonic), 0644)
		if err != nil {
			panic(err)
		}
	}
	mnemonic, err := os.ReadFile(mnemonicFile)
	if err != nil {
		panic(err)
	}
	wallet, err = hdwallet.NewFromMnemonic(string(mnemonic))
	if err != nil {
		panic(err)
	}
}

func GetTOwner(t *testing.T) *bind.TransactOpts {
	account, err := wallet.Derive(accounts.DefaultBaseDerivationPath, true)
	require.NoError(t, err)
	pk, err := wallet.PrivateKey(account)
	require.NoError(t, err)
	opts, err := bind.NewKeyedTransactorWithChainID(pk, ChainID)
	require.NoError(t, err)
	return opts
}

func GetTEoa(t *testing.T) *bind.TransactOpts {
	eoaTCount++
	account, err := wallet.Derive(append(accounts.DefaultRootDerivationPath, eoaTCount), true)
	require.NoError(t, err)
	pk, err := wallet.PrivateKey(account)
	require.NoError(t, err)
	opts, err := bind.NewKeyedTransactorWithChainID(pk, ChainID)
	require.NoError(t, err)
	return opts
}

func GetTEoas(t *testing.T, count int) []*bind.TransactOpts {
	opts := make([]*bind.TransactOpts, count)
	for i := 0; i < count; i++ {
		opts[i] = GetTEoa(t)
	}
	return opts
}

func GetEoaAt(chainID *big.Int, index uint32) (*bind.TransactOpts, error) {
	if account, err := wallet.Derive(append(accounts.DefaultRootDerivationPath, index), true); err != nil {
		return nil, err
	} else if pk, err := wallet.PrivateKey(account); err != nil {
		return nil, err
	} else {
		return bind.NewKeyedTransactorWithChainID(pk, chainID)
	}
}

func GetEoa(chainID *big.Int) (*bind.TransactOpts, error) {
	if account, err := wallet.Derive(append(accounts.DefaultRootDerivationPath, eoaCount), true); err != nil {
		return nil, err
	} else if pk, err := wallet.PrivateKey(account); err != nil {
		return nil, err
	} else {
		eoaCount++
		return bind.NewKeyedTransactorWithChainID(pk, chainID)
	}
}

func GetEoas(chainID *big.Int, count int) ([]*bind.TransactOpts, error) {
	var (
		opts []*bind.TransactOpts = make([]*bind.TransactOpts, count)
		err  error                = nil
	)
	cIndex := eoaCount
	for i := 0; i < count; i++ {
		if opts[i], err = GetEoa(chainID); err != nil {
			eoaCount = cIndex
			return nil, err
		}
	}
	return opts, err
}
