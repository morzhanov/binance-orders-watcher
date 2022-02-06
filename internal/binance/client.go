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
	GetAllOrdersForSymbol(symbol string) ([]*BinanceOrder, error)
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
	Price               string `json:"price"`
	OrigQty             string `json:"origQty"`
	ExecutedQty         string `json:"executedQty"`
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status              string `json:"status"`
	TimeInForce         string `json:"timeInForce"`
	Type                string `json:"type"`
	Side                string `json:"side"`
	StopPrice           string `json:"stopPrice"`
	IcebergQty          string `json:"icebergQty"`
	Time                int    `json:"time"`
	UpdateTime          int    `json:"updateTime"`
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
	if err != nil {
		return nil, err
	}
	req.Header[ApiKeyHeaderName] = []string{c.apiKey}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
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

func (c *client) GetAllOrdersForSymbol(symbol string) ([]*BinanceOrder, error) {
	ts := time.Now()
	query := fmt.Sprintf("timestamp=%d&recvWindow=10000&symbol=%s", ts.UnixMilli(), symbol)
	signature := c.createSignature(query)

	uri := fmt.Sprintf("%s/api/v3/allOrders?%s&signature=%s", c.prodURI, query, signature)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header[ApiKeyHeaderName] = []string{c.apiKey}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
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
	uri := fmt.Sprintf("%s/api/v3/ticker/price", c.prodURI)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
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
