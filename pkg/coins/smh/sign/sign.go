package sign

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/spacemesh-sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/spacemesh-sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/sol"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/spacemeshos/go-spacemesh/common/types"
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

	// just use in testnet
	types.DefaultTestAddressConfig()

	pubStr := signer.PublicKey().String()
	priStr := hex.EncodeToString(signer.PrivateKey())
	accStr := types.GenerateAddress([]byte(pubStr)).String()

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
	info := sol.SignMsgTx{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	var (
		from   = info.BaseInfo.From
		to     = info.BaseInfo.To
		value  = info.BaseInfo.Value
		rbHash = info.RecentBlockHash
	)

	fPublicKey, err := solana.PublicKeyFromBase58(from)
	if err != nil {
		return nil, err
	}

	tPublicKey, err := solana.PublicKeyFromBase58(to)
	if err != nil {
		return nil, err
	}

	rhash, err := solana.HashFromBase58(rbHash)
	if err != nil {
		return nil, err
	}

	lamports, accuracy := sol.ToLarm(value)
	if accuracy != big.Exact {
		log.Warnf("transafer sol amount not accuracy: from %v-> to %v", value, lamports)
	}

	// build tx
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				lamports,
				fPublicKey,
				tPublicKey,
			).Build(),
		},
		rhash,
		solana.TransactionPayer(fPublicKey),
	)
	if err != nil {
		return nil, err
	}

	pk, err := oss.GetObject(ctx, coins.GetS3KeyPrxfix(tokenInfo)+from, true)
	if err != nil {
		return nil, err
	}

	accountFrom := solana.PrivateKey(pk)
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if accountFrom.PublicKey().Equals(key) {
				return &accountFrom
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	err = tx.VerifySignatures()
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	if err := tx.MarshalWithEncoder(bin.NewBinEncoder(&buf)); err != nil {
		return nil, err
	}

	_out := sol.BroadcastRequest{
		Signature: buf.Bytes(),
	}

	return json.Marshal(_out)
}
