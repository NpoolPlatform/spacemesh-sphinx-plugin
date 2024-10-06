package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	chiaClient "github.com/NpoolPlatform/chia-client/pkg/client"
	"github.com/NpoolPlatform/chia-client/pkg/transaction"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/chia"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
)

// here register plugin func
func init() {
	register.RegisteTokenHandler(
		coins.Chia,
		register.OpGetBalance,
		walletBalance,
	)
	register.RegisteTokenHandler(
		coins.Chia,
		register.OpPreSign,
		preSign,
	)
	register.RegisteTokenHandler(
		coins.Chia,
		register.OpBroadcast,
		broadcast,
	)
	register.RegisteTokenHandler(
		coins.Chia,
		register.OpSyncTx,
		syncTx,
	)

	err := register.RegisteAbortFuncErr(sphinxplugin.CoinType_CoinTypechia, chia.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = register.RegisteAbortFuncErr(sphinxplugin.CoinType_CoinTypetchia, chia.TxFailErr)
	if err != nil {
		panic(err)
	}
}

func walletBalance(ctx context.Context, in []byte, _ *coins.TokenInfo) (out []byte, err error) {
	info := ct.WalletBalanceRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return in, err
	}

	v, ok := env.LookupEnv(env.ENVCOINNET)
	if !ok {
		return in, env.ErrEVNCoinNet
	}

	if !coins.CheckSupportNet(v) {
		return in, env.ErrEVNCoinNetValue
	}

	if info.Address == "" {
		return in, env.ErrAddressInvalid
	}

	cli := chia.Client()
	var amount uint64
	err = cli.WithClient(ctx, func(ctx context.Context, c *chiaClient.Client) (bool, error) {
		amount, err = c.GetBalance(ctx, info.Address)
		if err != nil {
			return true, err
		}

		return false, err
	})
	if err != nil {
		return in, err
	}

	balance, err := chia.ToXCH(amount)
	if err != nil {
		return in, err
	}
	f, exact := balance.Float64()
	if exact {
		log.Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
	}

	_out := ct.WalletBalanceResponse{
		Balance:    f,
		BalanceStr: balance.String(),
	}

	return json.Marshal(_out)
}

func preSign(ctx context.Context, in []byte, _ *coins.TokenInfo) (out []byte, err error) {
	info := ct.BaseInfo{}
	if err := json.Unmarshal(in, &info); err != nil {
		return in, err
	}

	if !coins.CheckSupportNet(info.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	amount := chia.ToMojo(info.Value)
	fee := uint64(chia.MinStandardTXFee)
	// todo: should check,maybe can caculate from chain
	var unsignedTx *transaction.UnsignedTx
	cli := chia.Client()
	err = cli.WithClient(ctx, func(ctx context.Context, c *chiaClient.Client) (bool, error) {
		unsignedTx, err = transaction.GenUnsignedTx(ctx, c, info.From, info.To, amount, fee)
		if err != nil {
			return false, fmt.Errorf("%v, %v", chia.ErrChiaGenTx, err)
		}
		return false, nil
	})
	if err != nil {
		return in, err
	}
	waitSignTX := chia.WaitSginTX{
		UnsignedTx: unsignedTx,
	}

	return json.Marshal(waitSignTX)
}

func broadcast(ctx context.Context, in []byte, _ *coins.TokenInfo) (out []byte, err error) {
	info := chia.BroadcastRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return in, err
	}

	client := chia.Client()
	var txid string
	err = client.WithClient(ctx, func(ctx context.Context, c *chiaClient.Client) (bool, error) {
		txid, err = c.PushTX(ctx, info.SpendBundle)
		if err != nil {
			return false, fmt.Errorf("%v, %v", chia.ErrChiaTxWrong, err)
		}
		return false, nil
	})
	if err != nil {
		return in, err
	}

	_out := chia.SyncRequest{
		TxID:         txid,
		SpentCoinIDs: info.SpentCoinIDs,
	}

	return json.Marshal(_out)
}

// syncTx sync transaction status on chain
func syncTx(ctx context.Context, in []byte, _ *coins.TokenInfo) (out []byte, err error) {
	info := chia.SyncRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return in, err
	}

	client := chia.Client()
	err = client.WithClient(ctx, func(ctx context.Context, c *chiaClient.Client) (bool, error) {
		inMemPool, err := c.CheckTxIDInMempool(ctx, info.TxID)
		if err != nil {
			return false, err
		}
		if inMemPool {
			return false, chia.ErrChiaTxSyncing
		}

		isSpent, err := c.CheckCoinsIsSpent(ctx, info.SpentCoinIDs)
		if err != nil {
			return false, err
		}
		if !isSpent {
			return false, chia.ErrChiaTxSyncing
		}
		return false, nil
	})
	if err != nil {
		return in, err
	}

	return in, nil
}
