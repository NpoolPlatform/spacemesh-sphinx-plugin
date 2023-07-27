package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/getter"
	coins_register "github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	pconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/types"
)

func init() {
	if err := register(
		"task::synctx",
		// TODO register not set duration
		3*time.Second,
		syncTxWorker,
	); err != nil {
		fatalf("task::synctx", "task already register")
	}
}

func calcDuration() time.Duration {
	du := config.GetENV().SyncInterval
	if du > 0 {
		return time.Duration(du) * time.Second
	}

	coinInfo, err := env.GetCoinInfo()
	if err != nil {
		panic(fmt.Sprintf("task::synctx failed to read %v, %v", env.ENVCOINTYPE, err))
	}

	coinType := coins.CoinStr2CoinType(coinInfo.NetworkType, coinInfo.CoinType)
	return coins.SyncTime[coinType]
}

func syncTxWorker(name string, _interval time.Duration) {
	interval := calcDuration()
	log.Infof("%v start,dispatch interval time: %v", name, interval.String())
	for range time.NewTicker(interval).C {
		func() {
			conn, err := client.GetGRPCConn(config.GetENV().Proxy)
			if err != nil {
				errorf(name, "call GetGRPCConn error: %v", err)
				return
			}

			coinInfo, err := env.GetCoinInfo()
			if err != nil {
				errorf(name, "get coin info from env error: %v", err)
				return
			}

			_coinType := coins.CoinStr2CoinType(coinInfo.NetworkType, coinInfo.CoinType)
			tState := sphinxproxy.TransactionState_TransactionStateSync

			pClient := sphinxproxy.NewSphinxProxyClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), getTransactionsTimeout)
			ctx = pconst.SetPluginInfo(ctx)
			defer cancel()

			transInfos, err := pClient.GetTransactions(ctx, &sphinxproxy.GetTransactionsRequest{
				ENV:              coinInfo.NetworkType,
				CoinType:         _coinType,
				TransactionState: tState,
			})
			if err != nil {
				errorf(name, "call Transaction error: %v", err)
				return
			}

			for _, transInfo := range transInfos.GetInfos() {
				syncTx(ctx, name, transInfo, pClient)
			}
		}()
	}
}

func syncTx(ctx context.Context, name string, transInfo *sphinxproxy.TransactionInfo, pClient sphinxproxy.SphinxProxyClient) {
	ctx, cancel := context.WithTimeout(ctx, updateTransactionsTimeout)
	defer cancel()

	now := time.Now()
	defer func() {
		infof(
			name,
			"plugin handle coinType: %v transaction type: %v id: %v use: %v",
			transInfo.GetName(),
			transInfo.GetTransactionState(),
			transInfo.GetTransactionID(),
			time.Since(now).String(),
		)
	}()

	var (
		syncInfo    = types.SyncResponse{}
		tState      = sphinxproxy.TransactionState_TransactionStateSync
		nextState   = sphinxproxy.TransactionState_TransactionStateDone
		tokenInfo   *coins.TokenInfo
		handler     coins_register.HandlerDef
		respPayload []byte
		err         error
	)

	tokenInfo = getter.GetTokenInfo(transInfo.GetName())
	if tokenInfo == nil {
		errorf(name, "cannot get token info :%v", transInfo.GetName())
		nextState = sphinxproxy.TransactionState_TransactionStateFail
		goto done
	}

	handler, err = getter.GetTokenHandler(tokenInfo.TokenType, coins_register.OpSyncTx)
	if err != nil {
		errorf(name, "cannot get token handler : %v - %v", tokenInfo.TokenType, coins_register.OpSyncTx)
		nextState = sphinxproxy.TransactionState_TransactionStateFail
		goto done
	}

	respPayload, err = handler(ctx, transInfo.GetPayload(), tokenInfo)
	if err == nil {
		goto done
	}
	if getter.Abort(tokenInfo.CoinType, err) {
		errorf(name,
			"sync transaction: %v error: %v stop",
			transInfo.GetTransactionID(),
			err,
		)
		nextState = sphinxproxy.TransactionState_TransactionStateFail
		goto done
	}

	errorf(name,
		"sync transaction: %v error: %v retry",
		transInfo.GetTransactionID(),
		err,
	)
	return

	// TODO: delete this dirty code
done:
	{
		if respPayload != nil {
			if err := json.Unmarshal(respPayload, &syncInfo); err != nil {
				errorf(name, "unmarshal sync info error: %v", err)
				return
			}
		}
	}

	if _, err := pClient.UpdateTransaction(ctx, &sphinxproxy.UpdateTransactionRequest{
		TransactionID:        transInfo.GetTransactionID(),
		TransactionState:     tState,
		NextTransactionState: nextState,
		ExitCode:             syncInfo.ExitCode,
		Payload:              respPayload,
	}); err != nil {
		errorf(name, "UpdateTransaction transaction: %v error: %v", transInfo.GetTransactionID(), err)
		return
	}

	infof(name, "UpdateTransaction transaction: %v done", transInfo.GetTransactionID())
}
