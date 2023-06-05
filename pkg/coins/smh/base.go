package smh

import (
	"errors"
	"math/big"
	"strings"

	v1 "github.com/NpoolPlatform/message/npool/basetypes/v1"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
)

const (
	// There are 10^12 SMIDGE in one SMH.
	SmidgePreSmh uint64 = 1000000000000

	ChainType       = sphinxplugin.ChainType_Spacemesh
	ChainNativeUnit = "SMH"
	ChainAtomicUnit = "SMD"
	ChainUnitExp    = 12
	//TODO:not sure,beacause the chain have no mainnet
	ChainID             = "1"
	ChainNickname       = "Spacemesh"
	ChainNativeCoinName = "spacemesh"
)

var (
	// EmptyWalletL ..
	EmptyWalletL = big.Int{}
	// EmptyWalletS ..
	EmptyWalletS = big.Float{}
)

var (
	// ErrSmhAddressWrong ..
	ErrSmhAddressWrong = errors.New("from or to address wrong")
	// ErrSmhNodeNotSynced ..
	ErrSmhNodeNotSynced = errors.New("spacemesh node not synced")
	// ErrSmhBlockNotFound ..
	ErrSmhBlockNotFound = errors.New("not found confirmed block in spacemesh chain")
	// ErrSmlSignatureWrong ..
	ErrSmlSignatureWrong = errors.New("spacemesh signature is wrong or failed")
	// ErrSmlTxWrong ..
	ErrSmlTxWrong = errors.New("spacemesh transaction is wrong or failed")
	// ErrSmlWaitSpawnFinish ..
	ErrSmlWaitSpawnFinish = errors.New("wait spwan transaction finish")
	// ErrSmlWaitSpendFinish ..
	ErrSmlWaitSpendFinish = errors.New("wait spend transaction finish")
)

var (
	lamportsLow = `Transfer: insufficient lamports`
	stopErrMsg  = []string{
		lamportsLow,
		ErrSmhAddressWrong.Error(),
		ErrSmhBlockNotFound.Error(),
		ErrSmlSignatureWrong.Error(),
		ErrSmlTxWrong.Error(),
		ErrSmhNodeNotSynced.Error(),
	}
	spacemeshToken = &coins.TokenInfo{OfficialName: "Spacemesh", Decimal: 12, Unit: "SMH", Name: ChainNativeCoinName, OfficialContract: ChainNativeCoinName, TokenType: coins.Spacemesh}
)

func init() {
	// set chain info
	spacemeshToken.ChainType = ChainType
	spacemeshToken.ChainNativeUnit = ChainNativeUnit
	spacemeshToken.ChainAtomicUnit = ChainAtomicUnit
	spacemeshToken.ChainUnitExp = ChainUnitExp
	spacemeshToken.GasType = v1.GasType_GasUnsupported
	spacemeshToken.ChainID = ChainID
	spacemeshToken.ChainNickname = ChainNickname
	spacemeshToken.ChainNativeCoinName = ChainNativeCoinName

	spacemeshToken.Waight = 100
	spacemeshToken.Net = coins.CoinNetMain
	spacemeshToken.Contract = spacemeshToken.OfficialContract
	spacemeshToken.CoinType = sphinxplugin.CoinType_CoinTypespacemesh
	register.RegisteTokenInfo(spacemeshToken)
}

func ToSmh(smidge uint64) *big.Float {
	// Convert lamports to SMH:
	return big.NewFloat(0).
		Quo(
			big.NewFloat(0).SetUint64(smidge),
			big.NewFloat(0).SetUint64(SmidgePreSmh),
		)
}

func ToSmidge(value float64) (uint64, big.Accuracy) {
	return big.NewFloat(0).Mul(
		big.NewFloat(0).SetFloat64(value),
		big.NewFloat(0).SetUint64(SmidgePreSmh),
	).Uint64()
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
