package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	smhclient "github.com/NpoolSpacemesh/spacemesh-plugin/client"
	v1 "github.com/spacemeshos/api/release/go/spacemesh/v1"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/smh"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
)

// here register plugin func
func init() {
	register.RegisteTokenHandler(
		coins.Spacemesh,
		register.OpGetBalance,
		walletBalance,
	)
	register.RegisteTokenHandler(
		coins.Spacemesh,
		register.OpPreSign,
		preSign,
	)
	register.RegisteTokenHandler(
		coins.Spacemesh,
		register.OpBroadcast,
		broadcast,
	)
	register.RegisteTokenHandler(
		coins.Spacemesh,
		register.OpSyncTx,
		syncTx,
	)

	err := register.RegisteAbortFuncErr(sphinxplugin.CoinType_CoinTypespacemesh, smh.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = register.RegisteAbortFuncErr(sphinxplugin.CoinType_CoinTypetspacemesh, smh.TxFailErr)
	if err != nil {
		panic(err)
	}
}

func walletBalance(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
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

	cli := smh.Client()
	var accountState *v1.Account
	err = cli.WithClient(ctx, func(_ctx context.Context, c *smhclient.Client) (bool, error) {
		accountState, err = c.AccountState(v1.AccountId{Address: info.Address})
		if err != nil || accountState == nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return in, err
	}

	balance := smh.ToSmh(accountState.StateProjected.GetBalance().GetValue())
	f, exact := balance.Float64()
	if exact != big.Exact {
		log.Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
	}

	_out := ct.WalletBalanceResponse{
		Balance:    f,
		BalanceStr: balance.String(),
	}

	return json.Marshal(_out)
}

func preSign(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	info := ct.BaseInfo{}
	if err := json.Unmarshal(in, &info); err != nil {
		return in, err
	}

	if !coins.CheckSupportNet(info.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	// todo: should check,maybe can caculate from chain
	gasPrice := uint64(1)
	nonce := uint64(0)
	genesisID := []byte{}

	client := smh.Client()
	err = client.WithClient(ctx, func(ctx context.Context, c *smhclient.Client) (bool, error) {
		accState, err := c.AccountState(v1.AccountId{Address: info.From})
		if err != nil {
			return true, err
		}
		nonce = accState.StateProjected.Counter
		genesisID, err = c.GetGenesisID()
		if err != nil {
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return in, err
	}

	_out := smh.SignMsgTx{
		BaseInfo:  info,
		GasPrice:  gasPrice,
		GenesisID: genesisID,
		Nonce:     nonce,
	}

	return json.Marshal(_out)
}

func broadcast(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	info := smh.BroadcastRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return in, err
	}

	client := smh.Client()
	if err != nil {
		return in, err
	}

	var txState *v1.TransactionState

	err = client.WithClient(ctx, func(ctx context.Context, c *smhclient.Client) (bool, error) {
		txState, err = c.SubmitCoinTransaction(info.TxData)
		if err != nil {
			return true, err
		}
		return false, nil
	})

	if err != nil {
		return in, err
	}

	_out := ct.SyncRequest{
		TxID: txState.Id.String(),
	}

	return json.Marshal(_out)
}

// syncTx sync transaction status on chain
func syncTx(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	info := ct.SyncRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return in, err
	}

	client := smh.Client()
	var txState *v1.TransactionState
	err = client.WithClient(ctx, func(ctx context.Context, c *smhclient.Client) (bool, error) {
		txState, _, err = c.TransactionState([]byte(info.TxID), false)
		if err != nil {
			return true, err
		}
		return false, nil
	})

	if err != nil {
		return in, err
	}

	// todo: should check
	if txState == nil || (txState.State == v1.TransactionState_TRANSACTION_STATE_MESH ||
		txState.State == v1.TransactionState_TRANSACTION_STATE_MEMPOOL) {
		return in, env.ErrWaitMessageOnChain
	}

	if txState != nil && (txState.State == v1.TransactionState_TRANSACTION_STATE_UNSPECIFIED ||
		txState.State == v1.TransactionState_TRANSACTION_STATE_REJECTED ||
		txState.State == v1.TransactionState_TRANSACTION_STATE_CONFLICTING ||
		txState.State == v1.TransactionState_TRANSACTION_STATE_INSUFFICIENT_FUNDS) {
		sResp := &ct.SyncResponse{}
		sResp.ExitCode = -1
		out, mErr := json.Marshal(sResp)
		if mErr != nil {
			return in, mErr
		}
		return out, fmt.Errorf("%v,%v", smh.SmhTransactionFailed, err)
	}

	if txState != nil && txState.State == v1.TransactionState_TRANSACTION_STATE_PROCESSED {
		sResp := &ct.SyncResponse{}
		sResp.ExitCode = 0
		out, err := json.Marshal(sResp)
		if err != nil {
			return in, err
		}
		return out, nil
	}

	return in, smh.ErrSmhBlockNotFound
}
