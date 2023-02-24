package smh

import (
	"errors"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/spacemesh-sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
)

const (
	// There are 10^12 SMIDGE in one SMH.
	SMIDGE_PER_SMH uint64 = 1000000000000
)

var (
	// EmptyWalletL ..
	EmptyWalletL = big.Int{}
	// EmptyWalletS ..
	EmptyWalletS = big.Float{}
)

var (
	// ErrSolBlockNotFound ..
	ErrSolBlockNotFound = errors.New("not found confirmed block in solana chain")
	// ErrSolSignatureWrong ..
	ErrSolSignatureWrong = errors.New("solana signature is wrong or failed")
)

var (
	SolTransactionFailed = `sol transaction failed`
	lamportsLow          = `Transfer: insufficient lamports`
	stopErrMsg           = []string{lamportsLow, SolTransactionFailed}
	spacemeshToken       = &coins.TokenInfo{OfficialName: "Spacemesh", Decimal: 12, Unit: "SMH", Name: "spacemesh", OfficialContract: "spacemesh", TokenType: coins.Spacemesh}
)

func init() {
	spacemeshToken.Waight = 100
	spacemeshToken.Net = coins.CoinNetMain
	spacemeshToken.Contract = spacemeshToken.OfficialContract
	spacemeshToken.CoinType = sphinxplugin.CoinType_CoinTypespacemesh
	register.RegisteTokenInfo(spacemeshToken)
}

func ToSmh(smidge uint64) *big.Float {
	// Convert lamports to sol:
	return big.NewFloat(0).
		Quo(
			big.NewFloat(0).SetUint64(smidge),
			big.NewFloat(0).SetUint64(SMIDGE_PER_SMH),
		)
}

func ToSmidge(value float64) (uint64, big.Accuracy) {
	return big.NewFloat(0).Mul(
		big.NewFloat(0).SetFloat64(value),
		big.NewFloat(0).SetUint64(SMIDGE_PER_SMH),
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
