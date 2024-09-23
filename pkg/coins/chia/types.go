package chia

import (
	"github.com/NpoolPlatform/chia-client/pkg/transaction"
	"github.com/chia-network/go-chia-libs/pkg/types"
)

type WaitSginTX struct {
	*transaction.UnsignedTx
}

type BroadcastRequest struct {
	SpentCoinIDs []string
	*types.SpendBundle
}

type SyncRequest struct {
	TxID         string `json:"tx_id"`
	SpentCoinIDs []string
}
