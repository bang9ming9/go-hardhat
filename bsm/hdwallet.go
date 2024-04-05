package bsm

import (
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

func GetOwner(t *testing.T) *bind.TransactOpts {
	account, err := wallet.Derive(accounts.DefaultBaseDerivationPath, true)
	require.NoError(t, err)
	pk, err := wallet.PrivateKey(account)
	require.NoError(t, err)
	opts, err := bind.NewKeyedTransactorWithChainID(pk, ChainID())
	require.NoError(t, err)
	return opts
}

func GetEOA(t *testing.T) *bind.TransactOpts {
	eoaCount++
	account, err := wallet.Derive(append(accounts.DefaultRootDerivationPath, eoaCount), true)
	require.NoError(t, err)
	pk, err := wallet.PrivateKey(account)
	require.NoError(t, err)
	opts, err := bind.NewKeyedTransactorWithChainID(pk, ChainID())
	require.NoError(t, err)
	return opts
}

func GetEOAs(t *testing.T, count int) []*bind.TransactOpts {
	opts := make([]*bind.TransactOpts, count)
	for i := range opts {
		opts[i] = GetEOA(t)
	}
	return opts
}
