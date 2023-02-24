package main

import (
	"context"
	"fmt"

	"github.com/NpoolPlatform/spacemesh-sphinx-plugin/pkg/coins/smh"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolSpacemesh/spacemesh-plugin/client"
	v1 "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func main() {
	config.SetENV(&config.ENVInfo{
		LocalWalletAddr:  "",
		PublicWalletAddr: "172.16.3.90:9092",
	})
	cli := smh.Client()
	cli.WithClient(context.Background(), func(ctx context.Context, c *client.Client) (bool, error) {
		fmt.Println(c.AccountState(v1.AccountId{Address: "stest1qqqqqqq28n6fw97jclu3tna6syxxy4elga2jtqgrf94zd"}))
		fmt.Println(c.NodeStatus())
		fmt.Println(smh.ToSmh(20554322965720), smh.SMIDGE_PER_SMH)
		return false, nil
	})
}
