package smh

import (
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/spacemeshos/go-spacemesh/common/types"
)

type SignMsgTx struct {
	BaseInfo  ct.BaseInfo `json:"base_info"`
	Nonce     uint64
	GasPrice  uint64
	Amount    uint64
	GenesisID []byte
}

type BroadcastRequest struct {
	SpawnTx *types.RawTx `json:"spawn_tx"`
	SpawnED bool
	SpendTx *types.RawTx `json:"spend_tx"`
}
