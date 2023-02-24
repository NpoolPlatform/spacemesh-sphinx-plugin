package task

import (
	"context"
	"encoding/json"
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
	// TODO: support from env or config dynamic set
	if err := register("task::nonce", 3*time.Second, nonceWorker); err != nil {
		fatalf("task::nonce", "task already register")
	}
}

func nonceWorker(name string, interval time.Duration) {
	log.Infof("%v start,dispatch interval time: %v", name, interval.String())
	for range time.NewTicker(interval).C {
		func() {
			conn, err := client.GetGRPCConn(config.GetENV().Proxy)
			if err != nil {
				errorf(name, "call GetGRPCConn error: %v", err)
				return
			}

			coinNetwork, coinType, err := env.CoinInfo()
			if err != nil {
				errorf(name, "get coin info from env error: %v", err)
				return
			}

			_coinType := coins.CoinStr2CoinType(coinNetwork, coinType)
			tState := sphinxproxy.TransactionState_TransactionStateWait
			pClient := sphinxproxy.NewSphinxProxyClient(conn)

			ctx, cancel := context.WithTimeout(context.Background(), getTransactionsTimeout)
			ctx = pconst.SetPluginInfo(ctx)
			defer cancel()

			transInfos, err := pClient.GetTransactions(ctx, &sphinxproxy.GetTransactionsRequest{
				ENV:              coinNetwork,
				CoinType:         _coinType,
				TransactionState: tState,
			})
			if err != nil {
				errorf(name, "call Transaction error: %v", err)
				return
			}

			for _, transInfo := range transInfos.GetInfos() {
				nonce(ctx, name, transInfo, pClient)
			}
		}()
	}
}

func nonce(ctx context.Context, name string, transInfo *sphinxproxy.TransactionInfo, pClient sphinxproxy.SphinxProxyClient) {
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
		tState         = sphinxproxy.TransactionState_TransactionStateWait
		nextState      = sphinxproxy.TransactionState_TransactionStateSign
		tokenInfo      *coins.TokenInfo
		handler        coins_register.HandlerDef
		respPayload    []byte
		preSignPayload []byte
		err            error
	)

	tokenInfo = getter.GetTokenInfo(transInfo.GetName())
	if tokenInfo == nil {
		nextState = sphinxproxy.TransactionState_TransactionStateFail
		goto done
	}

	handler, err = getter.GetTokenHandler(tokenInfo.TokenType, coins_register.OpPreSign)
	if err != nil {
		nextState = sphinxproxy.TransactionState_TransactionStateFail
		goto done
	}

	preSignPayload, err = json.Marshal(types.BaseInfo{
		ENV:      tokenInfo.Net,
		CoinType: tokenInfo.CoinType,
		From:     transInfo.GetFrom(),
		To:       transInfo.GetTo(),
		Value:    transInfo.GetAmount(),
	})
	if err != nil {
		errorf(name, "marshal presign info error: %v", err)
		return
	}
	respPayload, err = handler(ctx, preSignPayload, tokenInfo)
	if err == nil {
		goto done
	}

	if getter.Abort(tokenInfo.CoinType, err) {
		errorf(name,
			"pre sign transaction: %v error: %v stop",
			transInfo.GetTransactionID(),
			err,
		)
		nextState = sphinxproxy.TransactionState_TransactionStateFail
		goto done
	}

	errorf(name,
		"pre sign transaction: %v error: %v retry",
		transInfo.GetTransactionID(),
		err,
	)
	return

done:
	if _, err := pClient.UpdateTransaction(ctx, &sphinxproxy.UpdateTransactionRequest{
		TransactionID:        transInfo.GetTransactionID(),
		TransactionState:     tState,
		NextTransactionState: nextState,
		Payload:              respPayload,
	}); err != nil {
		errorf(name, "UpdateTransaction transaction: %v error: %v", transInfo.GetTransactionID(), err)
		return
	}

	infof(name, "UpdateTransaction transaction: %v done", transInfo.GetTransactionID())
}
