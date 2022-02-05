package client

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/morzhanov/binance-orders-watcher/internal/binance"
)

type Client interface{}

type client struct {
	binClient binance.Client
}

func New() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	http.Handle("/", r)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: 	1. add basic app auth or read auth token from headers
	//			2. add login for binance api or check auth token from headers
	//			3. load all orders
	//			4. load all prices for orders
	//			5. check alerts in database, if alerts are not configured configure them
	//				a. guess this step should be performed on the background and alerts should be sent to email
	//			6. render orders and gap between market price and order price
}
