package smh

import (
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
)

type SignMsgTx struct {
	BaseInfo  ct.BaseInfo `json:"base_info"`
	Nonce     uint64
	GasPrice  uint64
	Amount    uint64
	GenesisID []byte
}

type BroadcastRequest struct {
	TxData []byte `json:"signature"`
}
