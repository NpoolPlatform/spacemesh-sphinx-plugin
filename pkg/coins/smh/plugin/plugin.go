package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	smhclient "github.com/NpoolSpacemesh/spacemesh-plugin/client"
	v1 "github.com/spacemeshos/api/release/go/spacemesh/v1"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/spacemesh-sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/spacemesh-sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/spacemesh-sphinx-plugin/pkg/coins/smh"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
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

	err := register.RegisteAbortFuncErr(sphinxplugin.CoinType_CoinTypespacemesh, sol.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = register.RegisteAbortFuncErr(sphinxplugin.CoinType_CoinTypetspacemesh, sol.TxFailErr)
	if err != nil {
		panic(err)
	}
}

func W(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	return walletBalance(ctx, in, tokenInfo)
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

	client := sol.Client()

	var recentBlockHash *rpc.GetLatestBlockhashResult
	err = client.WithClient(ctx, func(_ctx context.Context, cli *rpc.Client) (bool, error) {
		recentBlockHash, err = cli.GetLatestBlockhash(_ctx, rpc.CommitmentFinalized)
		if err != nil || recentBlockHash == nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return in, err
	}

	_out := sol.SignMsgTx{
		BaseInfo:        info,
		RecentBlockHash: recentBlockHash.Value.Blockhash.String(),
	}

	return json.Marshal(_out)
}

func broadcast(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	info := sol.BroadcastRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return in, err
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(info.Signature))
	if err != nil {
		return in, err
	}

	err = tx.VerifySignatures()
	if err != nil {
		return in, sol.ErrSolSignatureWrong
	}

	client := sol.Client()
	if err != nil {
		return in, err
	}
	var cid solana.Signature
	err = client.WithClient(ctx, func(_ctx context.Context, cli *rpc.Client) (bool, error) {
		cid, err = cli.SendTransaction(_ctx, tx)
		if err != nil && !sol.TxFailErr(err) {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return in, err
	}

	_out := ct.SyncRequest{
		TxID: cid.String(),
	}

	return json.Marshal(_out)
}

// syncTx sync transaction status on chain
func syncTx(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	info := ct.SyncRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return in, err
	}

	signature, err := solana.SignatureFromBase58(info.TxID)
	if err != nil {
		return in, err
	}

	client := sol.Client()
	var chainMsg *rpc.GetTransactionResult
	err = client.WithClient(ctx, func(_ctx context.Context, cli *rpc.Client) (bool, error) {
		chainMsg, err = cli.GetTransaction(
			_ctx,
			signature,
			&rpc.GetTransactionOpts{
				Encoding:   solana.EncodingBase58,
				Commitment: rpc.CommitmentFinalized,
			})
		if err != nil {
			return true, err
		}
		return false, err
	})

	if err != nil {
		return in, err
	}

	if chainMsg == nil {
		return in, env.ErrWaitMessageOnChain
	}

	if chainMsg != nil && chainMsg.Meta.Err != nil {
		sResp := &ct.SyncResponse{}
		sResp.ExitCode = -1
		out, mErr := json.Marshal(sResp)
		if mErr != nil {
			return in, mErr
		}
		return out, fmt.Errorf("%v,%v", sol.SolTransactionFailed, err)
	}

	if chainMsg != nil && chainMsg.Meta.Err == nil {
		sResp := &ct.SyncResponse{}
		sResp.ExitCode = 0
		out, err := json.Marshal(sResp)
		if err != nil {
			return in, err
		}
		return out, nil
	}

	return in, sol.ErrSolBlockNotFound
}
