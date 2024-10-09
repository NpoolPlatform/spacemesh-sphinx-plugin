package sign

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NpoolPlatform/chia-client/pkg/account"
	"github.com/NpoolPlatform/chia-client/pkg/transaction"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/chia"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
)

func init() {
	register.RegisteTokenHandler(
		coins.Chia,
		register.OpWalletNew,
		createAccount,
	)
	register.RegisteTokenHandler(
		coins.Chia,
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

	isMainNet := info.ENV == coins.CoinNetMain

	acc, err := account.GenAccount()
	if err != nil {
		return nil, err
	}

	address, err := acc.GetAddress(isMainNet)
	if err != nil {
		return nil, err
	}

	_out := ct.NewAccountResponse{Address: address}
	out, err = json.Marshal(_out)
	if err != nil {
		return nil, err
	}

	skHex, err := acc.GetSKHex()
	if err != nil {
		return nil, err
	}

	err = oss.PutObject(ctx, coins.GetS3KeyPrxfix(tokenInfo)+address, []byte(skHex), true)
	if err != nil {
		return nil, err
	}

	return out, err
}

// signTx ..
func signTx(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	info := chia.WaitSginTX{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	sk, err := oss.GetObject(ctx, coins.GetS3KeyPrxfix(tokenInfo)+info.From, true)
	if err != nil {
		return nil, fmt.Errorf("%s, %s, address: %s", chia.ErrChiaAddressWrong, err, info.From)
	}

	spendBundle, err := transaction.GenSignedSpendBundle(info.UnsignedTx, string(sk))
	if err != nil {
		return nil, fmt.Errorf("%s, %s, address: %s", chia.ErrChiaSignatureWrong, err, info.From)
	}

	_out := chia.BroadcastRequest{
		SpendBundle:  spendBundle,
		SpentCoinIDs: info.SpentCoinIDs,
	}
	return json.Marshal(_out)
}
