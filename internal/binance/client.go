package binance

import (
	"fmt"
	"net/http"

	"github.com/morzhanov/binance-orders-watcher/internal/db"
)

type Client interface {
	GetOrders() ([]db.Order, error)
	GetPrices() ([]db.Price, error)
}

// TODO: 	1. find golang binance library or use http
//			2. find a way to authenticate users to perform binance api endpoint calls
//			3. implement Client interface api methods
type client struct {
	apiKey    string
	apiSecret string
	prodURI   string
}

func New(apiKey, apiSecret, prodURI string) Client {
	return &client{apiKey: apiKey, apiSecret: apiSecret, prodURI: prodURI}
}

func (c *client) GetOrders() {
	req, err := http.NewRequest(
		http.MethodGet,
		c.createURI("/api/v3/openOrders"),
		nil,
	)
	// TODO: add api key and secret headers to secrets or get access token
	//req.Header[]

	res, err := http.DefaultClient.Do(req)
	fmt.Printf("get orders req: res = %v, err = %v", res, err)
}

func (c *client) GetPrices(tickers []string) {
	for _, ticker := range tickers {
		c.getPrice(ticker)
	}
}

func (c *client) getPrice(ticker string) {
	req, err := http.NewRequest(
		http.MethodGet,
		c.createURI("/api/v3/ticker/price?symbol="+ticker),
		nil,
	)
	// TODO: add api key and secret headers to secrets or get access token
	//req.Header[]

	res, err := http.DefaultClient.Do(req)
	fmt.Printf("get ticker %s req: res = %v, err = %v", ticker, res, err)
}

func (c *client) createURI(suffix string) string {
	return c.prodURI + suffix
}
