package binance

type Client interface {
	GetOrders()
	GetPrices()
}

// TODO: 	1. find golang binance library or use http
//			2. find a way to authenticate users to perform binance api endpoint calls
//			3. implement Client interface api methods
type client struct {
}
