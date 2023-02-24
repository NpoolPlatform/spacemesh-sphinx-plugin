package sign

import (
	"context"
	"errors"
	"fmt"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
)

var (
	ErrCoinSignTypeAlreadyRegister = errors.New("coin sign type already register")
	ErrOpSignTypeAlreadyRegister   = errors.New("op sign type already register")

	ErrCoinSignTypeNotRegister = errors.New("coin sign type not register")
	ErrOpSignTypeNotRegister   = errors.New("op sign type not register")

	coinSignHandles       = make(map[sphinxplugin.CoinType]map[sphinxproxy.TransactionState]Handlef)
	coinWalletSignHandles = make(map[sphinxplugin.CoinType]map[sphinxproxy.TransactionType]Handlef)
)

type Handlef func(ctx context.Context, payload []byte) ([]byte, error)

func Register(coinType sphinxplugin.CoinType, opType sphinxproxy.TransactionState, handle Handlef) {
	coinPluginHandle, ok := coinSignHandles[coinType]
	if !ok {
		coinSignHandles[coinType] = make(map[sphinxproxy.TransactionState]Handlef)
	}
	if _, ok := coinPluginHandle[opType]; ok {
		panic(fmt.Errorf("coin type: %v for transaction: %v already registered", coinType, opType))
	}
	coinSignHandles[coinType][opType] = handle
}

func GetCoinSign(coinType sphinxplugin.CoinType, opType sphinxproxy.TransactionState) (Handlef, error) {
	// TODO: check nested map exist
	if _, ok := coinSignHandles[coinType]; !ok {
		return nil, ErrCoinSignTypeNotRegister
	}
	if _, ok := coinSignHandles[coinType][opType]; !ok {
		return nil, ErrOpSignTypeNotRegister
	}
	return coinSignHandles[coinType][opType], nil
}

func RegisterWallet(coinType sphinxplugin.CoinType, opType sphinxproxy.TransactionType, handle Handlef) {
	coinWalletPluginHandle, ok := coinWalletSignHandles[coinType]
	if !ok {
		coinWalletSignHandles[coinType] = make(map[sphinxproxy.TransactionType]Handlef)
	}
	if _, ok := coinWalletPluginHandle[opType]; ok {
		panic(fmt.Errorf("coin type: %v for transaction: %v already registered", coinType, opType))
	}
	coinWalletSignHandles[coinType][opType] = handle
}

func GetCoinWalletSign(coinType sphinxplugin.CoinType, opType sphinxproxy.TransactionType) (Handlef, error) {
	// TODO: check nested map exist
	if _, ok := coinWalletSignHandles[coinType]; !ok {
		return nil, ErrCoinSignTypeNotRegister
	}
	if _, ok := coinWalletSignHandles[coinType][opType]; !ok {
		return nil, ErrOpSignTypeNotRegister
	}
	return coinWalletSignHandles[coinType][opType], nil
}
