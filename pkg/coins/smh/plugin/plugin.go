package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/NpoolSpacemesh/spacemesh-plugin/account"
	smhclient "github.com/NpoolSpacemesh/spacemesh-plugin/client"
	v1 "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/common/types"

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

	if strings.HasPrefix(info.From, account.TestNet) {
		types.DefaultAddressConfig().NetworkHRP = account.TestNet
	} else {
		types.DefaultAddressConfig().NetworkHRP = account.MainNet
	}

	_, err = types.StringToAddress(info.From)
	if err != nil {
		return nil, fmt.Errorf("%s, %s, address: %s", smh.ErrSmhAddressWrong, err, info.From)
	}

	_, err = types.StringToAddress(info.To)
	if err != nil {
		return nil, fmt.Errorf("%s, %s, address: %s", smh.ErrSmhAddressWrong, err, info.To)
	}

	// todo: should check,maybe can caculate from chain
	gasPrice := uint64(2)
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

	// transfer hash32 address to hex
	// spacemesh accept address of format hash32,but spacemesh explore accept hex
	spendH32 := types.EmptyLayerHash
	spendH32.SetBytes(info.SpendTx.ID.Bytes())
	spendTxID := spendH32.Hex()

	var txState *v1.TransactionState

	client := smh.Client()
	err = client.WithClient(ctx, func(ctx context.Context, c *smhclient.Client) (bool, error) {
		if info.SpawnTx != nil {
			txState, err = c.SubmitCoinTransaction(info.SpawnTx.Raw)
			if err != nil {
				return true, nil
			}
			txState, tx, err := c.TransactionState(info.SpawnTx.ID[:], true)
			if err != nil {
				return true, nil
			}

			spawnH32 := types.EmptyLayerHash
			spawnH32.SetBytes(info.SpawnTx.ID.Bytes())
			spawnTxID := spawnH32.Hex()

			if txState.GetState() < v1.TransactionState_TRANSACTION_STATE_MEMPOOL || tx == nil {
				return false, fmt.Errorf("spawn tx %s failed, %s", spawnTxID, smh.ErrSmlTxWrong)
			}
			if txState.GetState() < v1.TransactionState_TRANSACTION_STATE_PROCESSED {
				return false, fmt.Errorf("spawn tx %s failed, %s", spawnTxID, smh.ErrSmlWaitSpawnFinish)
			}
			info.SpawnTx = nil
		}

		txState, err = c.SubmitCoinTransaction(info.SpendTx.Raw)
		if txState == nil {
			return true, nil
		}

		if err != nil {
			return true, fmt.Errorf("spend tx %s failed, %s", spendTxID, err)
		}

		return false, nil
	})

	if err != nil {
		return in, err
	}

	_out := ct.SyncRequest{
		TxID: spendTxID,
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
	var tx *v1.Transaction
	_txID := types.HexToHash32(info.TxID)
	err = client.WithClient(ctx, func(ctx context.Context, c *smhclient.Client) (bool, error) {
		txState, tx, err = c.TransactionState(_txID.Bytes(), true)
		if err != nil {
			return true, err
		}
		return false, nil
	})

	if err != nil {
		return in, err
	}

	if txState.GetState() < v1.TransactionState_TRANSACTION_STATE_MEMPOOL || tx == nil {
		return in, smh.ErrSmlTxWrong
	}

	if txState.GetState() < v1.TransactionState_TRANSACTION_STATE_PROCESSED {
		return in, smh.ErrSmlWaitSpendFinish
	}

	if txState.State == v1.TransactionState_TRANSACTION_STATE_PROCESSED {
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
