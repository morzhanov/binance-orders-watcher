package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/morzhanov/binance-orders-watcher/internal/db"
)

const (
	ApiKeyHeaderName = "X-MBX-APIKEY"
)

type Client interface {
	GetOrders() ([]*BinanceOrder, error)
	GetPrices() ([]*db.Price, error)
}

type client struct {
	apiKey    string
	apiSecret string
	prodURI   string
}

type BinanceOrder struct {
	Symbol              string `json:"symbol"`
	OrderId             int    `json:"orderId"`
	OrderListId         int    `json:"orderListId"`
	ClientOrderId       string `json:"clientOrderId"`
	Price               int    `json:"price"`
	OrigQty             int    `json:"origQty"`
	ExecutedQty         int    `json:"executedQty"`
	CummulativeQuoteQty int    `json:"cummulativeQuoteQty"`
	Status              string `json:"status"`
	TimeInForce         string `json:"timeInForce"`
	Type                string `json:"type"`
	Side                string `json:"side"`
	StopPrice           int    `json:"stopPrice"`
	IcebergQty          int    `json:"icebergQty"`
	Time                string `json:"time"`
	UpdateTime          string `json:"updateTime"`
	IsWorking           bool   `json:"isWorking"`
}

func New(apiKey, apiSecret, prodURI string) Client {
	return &client{apiKey: apiKey, apiSecret: apiSecret, prodURI: prodURI}
}

func (c *client) GetOrders() ([]*BinanceOrder, error) {
	ts := time.Now()
	query := fmt.Sprintf("timestamp=%d&recvWindow=10000", ts.UnixMilli())
	signature := c.createSignature(query)

	uri := fmt.Sprintf("%s/api/v3/openOrders?%s&signature=%s", c.prodURI, query, signature)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	req.Header[ApiKeyHeaderName] = []string{c.apiKey}
	res, err := http.DefaultClient.Do(req)
	defer res.Body.Close()

	ordersBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var orders []*BinanceOrder
	if err = json.Unmarshal(ordersBytes, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (c *client) GetPrices() ([]*db.Price, error) {
	uri := fmt.Sprintf("%sapi/v3/ticker/price")
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	res, err := http.DefaultClient.Do(req)
	priceBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var prices []*db.Price
	if err = json.Unmarshal(priceBytes, &prices); err != nil {
		return nil, err
	}
	return prices, nil
}

func (c *client) createSignature(text string) string {
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}
