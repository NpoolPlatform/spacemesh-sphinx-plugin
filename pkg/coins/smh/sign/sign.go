package sign

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/NpoolSpacemesh/spacemesh-plugin/account"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/smh"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk/wallet"
)

func init() {
	register.RegisteTokenHandler(
		coins.Spacemesh,
		register.OpWalletNew,
		createAccount,
	)
	register.RegisteTokenHandler(
		coins.Spacemesh,
		register.OpSign,
		signTx,
	)
}

func C(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	return createAccount(ctx, in, tokenInfo)
}

// createAccount ..
func createAccount(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	info := ct.NewAccountRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	if !coins.CheckSupportNet(info.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	var net string
	if info.ENV == coins.CoinNetMain {
		net = account.MainNet
	} else {
		net = account.TestNet
	}

	acc, err := account.CreateAccount(net)
	if err != nil {
		return nil, err
	}

	_out := ct.NewAccountResponse{Address: acc.Principal}
	out, err = json.Marshal(_out)
	if err != nil {
		return nil, err
	}

	err = oss.PutObject(ctx, coins.GetS3KeyPrxfix(tokenInfo)+acc.Principal, []byte(acc.Pri), true)
	if err != nil {
		return nil, err
	}

	return out, err
}

// signTx ..
func signTx(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	info := smh.SignMsgTx{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	if strings.HasPrefix(info.BaseInfo.To, account.TestNet) {
		types.DefaultAddressConfig().NetworkHRP = account.TestNet
	} else {
		types.DefaultAddressConfig().NetworkHRP = account.MainNet
	}

	toAddr, err := types.StringToAddress(info.BaseInfo.To)
	if err != nil {
		return nil, fmt.Errorf("%s, %s, address: %s", smh.ErrSmhAddressWrong, err, info.BaseInfo.To)
	}

	pk, err := oss.GetObject(ctx, coins.GetS3KeyPrxfix(tokenInfo)+info.BaseInfo.From, true)
	if err != nil {
		return nil, fmt.Errorf("%s, %s, address: %s", smh.ErrSmhAddressWrong, err, info.BaseInfo.From)
	}

	acc, err := account.CreateAccountFromHexPri(string(pk), "")
	if err != nil {
		return nil, fmt.Errorf("%s, %s, address: %s", smh.ErrSmhAddressWrong, err, info.BaseInfo.From)
	}

	signer := acc.GetSigner()

	amount, accuracy := smh.ToSmidge(info.BaseInfo.Value)

	if accuracy != big.Exact {
		log.Warnf("transafer spacemesh amount not accuracy: from %v-> to %v", info.BaseInfo.Value, amount)
	}

	_out := smh.BroadcastRequest{}

	spendTxNonce := info.Nonce
	if spendTxNonce == 0 {
		spawnTx := types.NewRawTx(
			wallet.SelfSpawn(
				signer.PrivateKey(),
				spendTxNonce,
				sdk.WithGenesisID(GenesisIDToH20(info.GenesisID)),
				sdk.WithGasPrice(info.GasPrice)))
		_out.SpawnTx = &spawnTx
		spendTxNonce++
	}

	spendTx := types.NewRawTx(
		wallet.Spend(
			signer.PrivateKey(),
			toAddr,
			amount,
			spendTxNonce,
			sdk.WithGenesisID(GenesisIDToH20(info.GenesisID)),
			sdk.WithGasPrice(info.GasPrice)))

	_out.SpendTx = &spendTx

	return json.Marshal(_out)
}

func GenesisIDToH20(genesisID []byte) types.Hash20 {
	_genesisID := types.EmptyLayerHash
	_genesisID.SetBytes(genesisID)
	h20 := types.Hash20{}
	copy(h20[:], _genesisID[12:])
	return h20
}
