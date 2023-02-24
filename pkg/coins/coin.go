package coins

import (
	"fmt"
	"strings"
	"time"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/utils"
)

type (
	TokenType string
)

const (
	Ethereum TokenType = "ethereum"
	Erc20    TokenType = "erc20"
	Erc721   TokenType = "erc721"
	// TODO: will remove,this type is for compatibility
	USDC TokenType = "usdc"

	Solana   TokenType = "solana"
	Bitcoin  TokenType = "bitcoin"
	Filecoin TokenType = "filecoin"

	Tron  TokenType = "tron"
	Trc20 TokenType = "trc20"

	Binancecoin TokenType = "binancecoin"
	Bep20       TokenType = "bep20"

	Spacemesh TokenType = "Spacemesh"
)

type TokenInfo struct {
	OfficialName     string
	OfficialContract string
	Contract         string // if ENV is main Contract = OfficialContract
	TokenType        TokenType
	Net              string
	Unit             string
	Decimal          int
	Name             string
	Waight           int
	DisableRegiste   bool
	CoinType         sphinxplugin.CoinType
}

const (
	CoinNetMain = "main"
	CoinNetTest = "test"
	TestPrefix  = "t"
)

var (
	// not export
	netCoinMap = map[string]map[string]sphinxplugin.CoinType{
		CoinNetMain: {
			"filecoin":    sphinxplugin.CoinType_CoinTypefilecoin,
			"bitcoin":     sphinxplugin.CoinType_CoinTypebitcoin,
			"ethereum":    sphinxplugin.CoinType_CoinTypeethereum,
			"usdterc20":   sphinxplugin.CoinType_CoinTypeusdterc20,
			"spacemesh":   sphinxplugin.CoinType_CoinTypespacemesh,
			"solana":      sphinxplugin.CoinType_CoinTypesolana,
			"usdttrc20":   sphinxplugin.CoinType_CoinTypeusdttrc20,
			"tron":        sphinxplugin.CoinType_CoinTypetron,
			"binancecoin": sphinxplugin.CoinType_CoinTypebinancecoin,
			"binanceusd":  sphinxplugin.CoinType_CoinTypebinanceusd,
			"usdcerc20":   sphinxplugin.CoinType_CoinTypeusdcerc20,
		},
		CoinNetTest: {
			"filecoin":    sphinxplugin.CoinType_CoinTypetfilecoin,
			"bitcoin":     sphinxplugin.CoinType_CoinTypetbitcoin,
			"ethereum":    sphinxplugin.CoinType_CoinTypetethereum,
			"usdterc20":   sphinxplugin.CoinType_CoinTypetusdterc20,
			"spacemesh":   sphinxplugin.CoinType_CoinTypetspacemesh,
			"solana":      sphinxplugin.CoinType_CoinTypetsolana,
			"usdttrc20":   sphinxplugin.CoinType_CoinTypetusdttrc20,
			"tron":        sphinxplugin.CoinType_CoinTypettron,
			"binancecoin": sphinxplugin.CoinType_CoinTypetbinancecoin,
			"binanceusd":  sphinxplugin.CoinType_CoinTypetbinanceusd,
			"usdcerc20":   sphinxplugin.CoinType_CoinTypetusdcerc20,
		},
	}

	// in order to compatible
	S3KeyPrxfixMap = map[string]string{
		"filecoin":     "filecoin/",
		"tfilecoin":    "filecoin/",
		"bitcoin":      "bitcoin/",
		"tbitcoin":     "bitcoin/",
		"ethereum":     "ethereum/",
		"tethereum":    "ethereum/",
		"usdterc20":    "ethereum/",
		"tusdterc20":   "ethereum/",
		"spacemesh":    "spacemesh/",
		"tspacemesh":   "spacemesh/",
		"solana":       "solana/",
		"tsolana":      "solana/",
		"usdttrc20":    "usdttrc20/",
		"tusdttrc20":   "usdttrc20/",
		"tron":         "tron/",
		"ttron":        "tron/",
		"binancecoin":  "binancecoin/",
		"tbinancecoin": "binancecoin/",
		"binanceusd":   "binanceusd/",
		"tbinanceusd":  "binanceusd/",
		"usdcerc20":    "usdcerc20/",
		"tusdcerc20":   "usdcerc20/",
	}

	// default sync time for waitting transaction on chain
	SyncTime = map[sphinxplugin.CoinType]time.Duration{
		sphinxplugin.CoinType_CoinTypefilecoin:  time.Second * 20,
		sphinxplugin.CoinType_CoinTypetfilecoin: time.Second * 20,

		sphinxplugin.CoinType_CoinTypebitcoin:  time.Minute * 7,
		sphinxplugin.CoinType_CoinTypetbitcoin: time.Minute * 7,

		sphinxplugin.CoinType_CoinTypeethereum:  time.Second * 12,
		sphinxplugin.CoinType_CoinTypetethereum: time.Second * 3,

		sphinxplugin.CoinType_CoinTypeusdterc20:  time.Second * 12,
		sphinxplugin.CoinType_CoinTypetusdterc20: time.Second * 3,

		sphinxplugin.CoinType_CoinTypeusdcerc20:  time.Second * 12,
		sphinxplugin.CoinType_CoinTypetusdcerc20: time.Second * 3,

		sphinxplugin.CoinType_CoinTypespacemesh:  time.Second * 30,
		sphinxplugin.CoinType_CoinTypetspacemesh: time.Second * 30,

		sphinxplugin.CoinType_CoinTypesolana:  time.Second * 1,
		sphinxplugin.CoinType_CoinTypetsolana: time.Second * 1,

		sphinxplugin.CoinType_CoinTypeusdttrc20:  time.Second * 2,
		sphinxplugin.CoinType_CoinTypetusdttrc20: time.Second * 2,

		sphinxplugin.CoinType_CoinTypetron:  time.Second * 2,
		sphinxplugin.CoinType_CoinTypettron: time.Second * 2,

		sphinxplugin.CoinType_CoinTypebinancecoin:  time.Second * 4,
		sphinxplugin.CoinType_CoinTypetbinancecoin: time.Second * 4,

		sphinxplugin.CoinType_CoinTypebinanceusd:  time.Second * 4,
		sphinxplugin.CoinType_CoinTypetbinanceusd: time.Second * 4,
	}
)

// CoinInfo report coin info
type CoinInfo struct {
	ENV      string // main or test
	Unit     string
	IP       string // wan ip
	Location string
}

// CheckSupportNet ..
func CheckSupportNet(netEnv string) bool {
	return (netEnv == CoinNetMain ||
		netEnv == CoinNetTest)
}

// TODO match case elegant deal
func CoinStr2CoinType(netEnv, coinStr string) sphinxplugin.CoinType {
	_netEnv := strings.ToLower(netEnv)
	_coinStr := strings.ToLower(coinStr)
	return netCoinMap[_netEnv][_coinStr]
}

func ToTestCoinType(coinType sphinxplugin.CoinType) sphinxplugin.CoinType {
	if coinType == sphinxplugin.CoinType_CoinTypeUnKnow {
		return sphinxplugin.CoinType_CoinTypeUnKnow
	}
	name := utils.ToCoinName(coinType)
	return CoinStr2CoinType(CoinNetTest, name)
}

func GetS3KeyPrxfix(tokenInfo *TokenInfo) string {
	if val, ok := S3KeyPrxfixMap[tokenInfo.Name]; ok {
		return val
	}

	name := tokenInfo.Name
	if tokenInfo.Net == CoinNetTest {
		name = strings.TrimPrefix(name, TestPrefix)
	}
	return fmt.Sprintf("%v/", name)
}

func GenerateName(tokenInfo *TokenInfo) string {
	chainType := utils.ToCoinName(tokenInfo.CoinType)
	name := strings.Trim(tokenInfo.OfficialName, " ")
	name = strings.ReplaceAll(name, " ", "-")
	return fmt.Sprintf("%v_%v_%v", chainType, tokenInfo.TokenType, name)
}

func GetChainType(in string) string {
	ret := strings.Split(in, "_")
	return ret[0]
}
