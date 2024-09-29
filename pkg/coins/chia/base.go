package chia

import (
	"errors"
	"math/big"
	"math"
	"strings"

	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/shopspring/decimal"
)

const (
	// There are 10^12 mojo in one XCH.
	ChiaExp int32 = -12

	ChainType           = sphinxplugin.ChainType_Chia
	ChainNativeUnit     = "XCH"
	ChainAtomicUnit     = "mojo"
	ChainUnitExp        = 12
	ChainID             = "mainnet"
	ChainNativeCoinName = "chia"

	// 0.00009 xch
	// MinStandardTXFee = 90000000
	// but in reality,it is usually set to 1
	MinStandardTXFee = 1
)

var (
	// EmptyWalletL ..
	EmptyWalletL = big.Int{}
	// EmptyWalletS ..
	EmptyWalletS = big.Float{}
)

var (
	// ErrChiaAddressWrong ..
	ErrChiaAddressWrong = errors.New("from or to address wrong")
	// ErrChiaGenTx ..
	ErrChiaGenTx = errors.New("falied to construct transaction")
	// ErrChiaNodeNotSynced ..
	ErrChiaNodeNotSynced = errors.New("chia node not synced")
	// ErrChiaSignatureWrong ..
	ErrChiaSignatureWrong = errors.New("chia signature is wrong or failed")
	// ErrChiaTxWrong ..
	ErrChiaTxWrong = errors.New("chia transaction is wrong or failed")
	// ErrChiaTxSyncing ..
	ErrChiaTxSyncing = errors.New("chia transaction is syncing")
)

var (
	lamportsLow = `Transfer: insufficient lamports`
	stopErrMsg  = []string{
		lamportsLow,
		ErrChiaAddressWrong.Error(),
		ErrChiaGenTx.Error(),
		ErrChiaSignatureWrong.Error(),
		ErrChiaTxWrong.Error(),
		ErrChiaNodeNotSynced.Error(),
	}
	chiaToken = &coins.TokenInfo{OfficialName: "chia", Decimal: ChainUnitExp, Unit: "Chia", Name: ChainNativeCoinName, OfficialContract: ChainNativeCoinName, TokenType: coins.Chia}
)

func init() {
	// set chain info
	chiaToken.ChainType = ChainType
	chiaToken.ChainNativeUnit = ChainNativeUnit
	chiaToken.ChainAtomicUnit = ChainAtomicUnit
	chiaToken.ChainUnitExp = ChainUnitExp
	chiaToken.GasType = basetypes.GasType_GasUnsupported
	chiaToken.ChainID = ChainID
	chiaToken.ChainNickname = ChainType.String()
	chiaToken.ChainNativeCoinName = ChainNativeCoinName

	chiaToken.Waight = 100
	chiaToken.Net = coins.CoinNetMain
	chiaToken.Contract = chiaToken.OfficialContract
	chiaToken.CoinType = sphinxplugin.CoinType_CoinTypechia
	register.RegisteTokenInfo(chiaToken)
}

func ToXCH(mojo uint64) (decimal.Decimal, error) {
	n, err := decimal.NewFromString(str)
	if err != nil {
		return decimal.NewFromInt(0), err
	}
	p, err := decimal.NewFromString(str)
	if err != nil {
		return decimal.NewFromInt(0), err
	}
	return n.Div(p), nil
}

func ToMojo(value float64) uint64 {
	return decimal.NewFromFloat(value).Shift(-ChiaExp).BigInt().Uint64()
}

func TxFailErr(err error) bool {
	if err == nil {
		return false
	}

	for _, v := range stopErrMsg {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}
