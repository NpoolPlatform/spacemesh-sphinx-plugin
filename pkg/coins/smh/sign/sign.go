package sign

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/spacemeshos/go-scale"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk"
	tplWallet "github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/spacemeshos/go-spacemesh/hash"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/smh"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/signing"
)

func init() {
	register.RegisteTokenHandler(
		coins.Spacemesh,
		register.OpWalletNew,
		createAccount,
	)
	register.RegisteTokenHandler(
		coins.Spacemesh,
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

	// if account equal nil will panic
	signer, err := signing.NewEdSigner()
	if err != nil {
		fmt.Println(err)
	}

	if info.ENV != coins.CoinNetMain {
		types.DefaultTestAddressConfig()
	} else {
		types.DefaultAddressConfig()
	}

	pubStr := signer.PublicKey().String()
	priStr := hex.EncodeToString(signer.PrivateKey())
	accStr := types.GenerateAddress([]byte(pubStr)).String()

	fmt.Println(priStr)
	fmt.Println(pubStr)
	fmt.Println(accStr)
	_out := ct.NewAccountResponse{
		Address: accStr,
	}

	out, err = json.Marshal(_out)
	if err != nil {
		return nil, err
	}

	err = oss.PutObject(ctx, coins.GetS3KeyPrxfix(tokenInfo)+accStr, []byte(priStr), true)
	if err != nil {
		return nil, err
	}

	return out, err
}

// signTx ..
func signTx(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	info := smh.SignMsgTx{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	pk, err := oss.GetObject(ctx, coins.GetS3KeyPrxfix(tokenInfo)+info.BaseInfo.From, true)
	if err != nil {
		return nil, err
	}
	signer, err := signing.NewEdSigner(signing.WithPrivateKey(pk))
	if err != nil {
		return nil, err
	}

	toAddr, err := types.StringToAddress(info.BaseInfo.To)
	if err != nil {
		return nil, err
	}
	amount, accuracy := smh.ToSmidge(info.BaseInfo.Value)

	if accuracy != big.Exact {
		log.Warnf("transafer spacemesh amount not accuracy: from %v-> to %v", info.BaseInfo.Value, amount)
	}

	payload := core.Payload{}
	payload.GasPrice = info.GasPrice
	payload.Nonce = info.Nonce

	args := tplWallet.SpendArguments{}
	args.Destination = toAddr
	args.Amount = amount

	tx := &types.Transaction{TxHeader: &types.TxHeader{}}
	spawnargs := tplWallet.SpawnArguments{}
	copy(spawnargs.PublicKey[:], signer.PublicKey().PublicKey)
	principal := core.ComputePrincipal(tplWallet.TemplateAddress, &spawnargs)

	_tx := encode(&sdk.TxVersion, &principal, &sdk.MethodSpend, &payload, &args)
	hh := hash.Sum(info.GenesisID, _tx)
	sig := ed25519.Sign(signer.PrivateKey(), hh[:])
	_tx = append(_tx, sig...)
	tx.RawTx = types.NewRawTx(_tx)
	tx.MaxSpend = 1

	// todo: not sure,please check
	// serializedTx, err := codec.Encode(tx)

	_out := smh.BroadcastRequest{
		TxData: tx.Raw,
	}

	return json.Marshal(_out)
}

func encode(fields ...scale.Encodable) []byte {
	buf := bytes.NewBuffer(nil)
	encoder := scale.NewEncoder(buf)
	for _, field := range fields {
		_, err := field.EncodeScale(encoder)
		if err != nil {
			panic(err)
		}
	}
	return buf.Bytes()
}
