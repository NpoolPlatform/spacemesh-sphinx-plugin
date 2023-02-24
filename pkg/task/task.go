package task

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/client"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"

	"github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/getter"
	coins_register "github.com/NpoolPlatform/sphinx-plugin-p2/pkg/coins/register"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/rpc"
	"google.golang.org/grpc"
)

var (
	chanBuff             = 1000
	delayDuration        = time.Second * 2
	registerCoinDuration = time.Second * 5
)

type pluginClient struct {
	closeBadConn chan struct{}
	exitChan     chan struct{}
	sendChannel  chan *sphinxproxy.ProxyPluginResponse

	once        sync.Once
	conn        *grpc.ClientConn
	proxyClient sphinxproxy.SphinxProxy_ProxyPluginClient
}

func Plugin(exitSig chan os.Signal, cleanChan chan struct{}) {
	newClient(exitSig, cleanChan)
}

func newClient(exitSig chan os.Signal, cleanChan chan struct{}) {
	proxyClient := &pluginClient{
		closeBadConn: make(chan struct{}),
		exitChan:     make(chan struct{}),
		sendChannel:  make(chan *sphinxproxy.ProxyPluginResponse, chanBuff),
	}

	conn, pc, err := proxyClient.newProxyClient()
	if err != nil {
		log.Errorf("create new proxy client error: %v", err)
		delayNewClient(exitSig, cleanChan)
		return
	}

	proxyClient.conn, proxyClient.proxyClient = conn, pc

	go proxyClient.watch(exitSig, cleanChan)
	go proxyClient.register()
	go proxyClient.send()
	go proxyClient.recv()
}

func delayNewClient(exitSig chan os.Signal, cleanChan chan struct{}) {
	time.Sleep(delayDuration)
	go newClient(exitSig, cleanChan)
}

func (c *pluginClient) closeProxyClient() {
	c.once.Do(func() {
		log.Info("close proxy conn and client")
		if c != nil {
			close(c.exitChan)
			if c.proxyClient != nil {
				if err := c.proxyClient.CloseSend(); err != nil {
					log.Warnf("close proxy conn and client error: %v", err)
				}
			}
			if c.conn != nil {
				if err := c.conn.Close(); err != nil {
					log.Warnf("close conn error: %v", err)
				}
			}
		}
	})
}

func (c *pluginClient) newProxyClient() (*grpc.ClientConn, sphinxproxy.SphinxProxy_ProxyPluginClient, error) {
	log.Info("start new proxy client")
	conn, err := client.GetGRPCConn(config.GetENV().Proxy)
	if err != nil {
		log.Errorf("call GetGRPCConn error: %v", err)
		return nil, nil, err
	}

	pClient := sphinxproxy.NewSphinxProxyClient(conn)
	proxyClient, err := pClient.ProxyPlugin(context.Background())
	if err != nil {
		log.Errorf("call Transaction error: %v", err)
		return nil, nil, err
	}

	log.Info("start new proxy client ok")
	return conn, proxyClient, nil
}

func (c *pluginClient) watch(exitSig chan os.Signal, cleanChan chan struct{}) {
	for {
		select {
		case <-c.closeBadConn:
			logger.Sugar().Info("start watch proxy client")
			<-c.closeBadConn
			c.closeProxyClient()
			logger.Sugar().Info("start watch proxy client exit")
			delayNewClient(exitSig, cleanChan)
		case <-exitSig:
			c.closeProxyClient()
			close(cleanChan)
			return
		}
	}
}

func (c *pluginClient) register() {
	for {
		select {
		case <-c.exitChan:
			log.Info("register new coin exit")
			return
		case <-time.After(registerCoinDuration):
			coinNetwork, _coinType, err := env.CoinInfo()
			if err != nil {
				log.Errorf("register new coin error: %v", err)
				continue
			}
			coinType := coins.CoinStr2CoinType(coinNetwork, _coinType)

			tokenInfos := getter.GetTokenInfos(coinType)

			tokensLen := 0
			// TODO: send a msg,which contain all tokentype bellow this plugin
			for _, tokenInfo := range tokenInfos {
				if tokenInfo.DisableRegiste {
					continue
				}
				resp := &sphinxproxy.ProxyPluginResponse{
					CoinType:        tokenInfo.CoinType,
					Name:            tokenInfo.Name,
					TransactionType: sphinxproxy.TransactionType_RegisterCoin,
					ENV:             tokenInfo.Net,
					Unit:            tokenInfo.Unit,
					PluginWanIP:     config.GetENV().WanIP,
					PluginPosition:  config.GetENV().Position,
				}
				tokensLen++
				c.sendChannel <- resp
			}
			log.Infof("register new coin: %v for %s network,has %v tokens,registered %v", coinType, coinNetwork, len(tokenInfos), tokensLen)
		}
	}
}

func (c *pluginClient) recv() {
	log.Info("plugin client start recv")
	for {
		select {
		case <-c.exitChan:
			log.Info("plugin client start recv exit")
			return
		default:
			req, err := c.proxyClient.Recv()
			if err != nil {
				log.Errorf("receiver info error: %v", err)
				if rpc.CheckCode(err) {
					c.closeBadConn <- struct{}{}
					break
				}
			}

			go func() {
				coinType := req.GetCoinType()
				transactionType := req.GetTransactionType()
				transactionID := req.GetTransactionID()

				log.Infof(
					"sphinx plugin recv info TransactionID: %v CoinType: %v TransactionType: %v",
					transactionID,
					coinType,
					transactionType,
				)

				now := time.Now()
				defer func() {
					log.Infof(
						"plugin handle coinType: %v transaction type: %v id: %v use: %v",
						coinType,
						transactionType,
						transactionID,
						time.Since(now).String(),
					)
				}()

				var resp *sphinxproxy.ProxyPluginResponse
				var err error
				var handler coins_register.HandlerDef
				// handler, err := coins.GetCoinBalancePlugin(coinType, transactionType)
				tokenInfo := getter.GetTokenInfo(req.Name)
				if tokenInfo == nil {
					log.Errorf("GetCoinPlugin get handler error: %v", err)
					resp = &sphinxproxy.ProxyPluginResponse{
						TransactionType: req.GetTransactionType(),
						CoinType:        req.GetCoinType(),
						TransactionID:   req.GetTransactionID(),
						RPCExitMessage:  err.Error(),
					}
					goto send
				}

				handler, err = getter.GetTokenHandler(tokenInfo.TokenType, coins_register.OpGetBalance)
				if err != nil {
					log.Errorf("GetCoinPlugin get handler error: %v", err)
					resp = &sphinxproxy.ProxyPluginResponse{
						TransactionType: req.GetTransactionType(),
						CoinType:        req.GetCoinType(),
						TransactionID:   req.GetTransactionID(),
						RPCExitMessage:  err.Error(),
					}
					goto send
				}
				{
					respPayload, err := handler(context.Background(), req.GetPayload(), tokenInfo)
					fmt.Println(tokenInfo)
					if err != nil {
						log.Errorf("GetCoinPlugin handle deal transaction error: %v", err)
						resp = &sphinxproxy.ProxyPluginResponse{
							TransactionType: req.GetTransactionType(),
							CoinType:        req.GetCoinType(),
							TransactionID:   req.GetTransactionID(),
							RPCExitMessage:  err.Error(),
						}
						goto send
					}

					resp = &sphinxproxy.ProxyPluginResponse{
						TransactionType: req.GetTransactionType(),
						CoinType:        req.GetCoinType(),
						TransactionID:   req.GetTransactionID(),
						Payload:         respPayload,
					}
				}

			send:
				c.sendChannel <- resp
			}()
		}
	}
}

func (c *pluginClient) send() {
	log.Info("plugin client start send")
	for {
		select {
		case <-c.exitChan:
			log.Info("plugin client start send exit")
			return
		case resp := <-c.sendChannel:
			err := c.proxyClient.Send(resp)
			if err != nil {
				log.Errorf("send info error: %v", err)
				if rpc.CheckCode(err) {
					c.closeBadConn <- struct{}{}
				}
			}
		}
	}
}
