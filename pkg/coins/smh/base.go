package smh

import (
	"errors"
	"math/big"
	"strings"

	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/shopspring/decimal"
)

const (
	// There are 10^9 SMIDGE in one SMH.
	SmhExp int32 = -9

	ChainType       = sphinxplugin.ChainType_Spacemesh
	ChainNativeUnit = "SMH"
	ChainAtomicUnit = "SMD"
	ChainUnitExp    = 9
	// TODO:not sure,beacause the chain have no mainnet
	ChainID             = "N/A"
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
	// ErrSmhInsufficient ..
	ErrSmhInsufficient = errors.New("the balance of from-address is insufficient to pay amount and gas")
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
		ErrSmhInsufficient.Error(),
		ErrSmhBlockNotFound.Error(),
		ErrSmlSignatureWrong.Error(),
		ErrSmlTxWrong.Error(),
		ErrSmhNodeNotSynced.Error(),
	}
	spacemeshToken = &coins.TokenInfo{OfficialName: "Spacemesh", Decimal: ChainUnitExp, Unit: "SMH", Name: ChainNativeCoinName, OfficialContract: ChainNativeCoinName, TokenType: coins.Spacemesh}
)

func init() {
	// set chain info
	spacemeshToken.ChainType = ChainType
	spacemeshToken.ChainNativeUnit = ChainNativeUnit
	spacemeshToken.ChainAtomicUnit = ChainAtomicUnit
	spacemeshToken.ChainUnitExp = ChainUnitExp
	spacemeshToken.GasType = basetypes.GasType_GasUnsupported
	spacemeshToken.ChainID = ChainID
	spacemeshToken.ChainNickname = ChainType.String()
	spacemeshToken.ChainNativeCoinName = ChainNativeCoinName

	spacemeshToken.Waight = 100
	spacemeshToken.Net = coins.CoinNetMain
	spacemeshToken.Contract = spacemeshToken.OfficialContract
	spacemeshToken.CoinType = sphinxplugin.CoinType_CoinTypespacemesh
	register.RegisteTokenInfo(spacemeshToken)
}

func ToSmh(smidge uint64) decimal.Decimal {
	return decimal.NewFromBigInt(big.NewInt(int64(smidge)), SmhExp)
}

func ToSmidge(value float64) uint64 {
	return decimal.NewFromFloat(value).Shift(-SmhExp).BigInt().Uint64()
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
