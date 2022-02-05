package cron

import (
	"log"
	"time"

	"github.com/morzhanov/binance-orders-watcher/internal/binance"
	"github.com/morzhanov/binance-orders-watcher/internal/db"
)

const (
	Interval = time.Minute * 30
)

type Cron interface {
	Run() error
}

type cronImp struct {
	binClient binance.Client
	db        db.Client
}

func New(binClient binance.Client, dbClient db.Client) Cron {
	return &cronImp{binClient: binClient, db: dbClient}
}

func (c *cronImp) Run() error {
	for {
		orders, err := c.binClient.GetOrders()
		if err != nil {
			log.Println("failed to get orders from binance: ", err)
			continue
		}
		prices, err := c.binClient.GetPrices()
		if err != nil {
			log.Println("failed to get prices from binance: ", err)
			continue
		}

		if err = c.db.SetOrders(orders); err != nil {
			log.Println("failed to set orders from binance to db: ", err)
			continue
		}
		if err = c.db.SetPrices(prices); err != nil {
			log.Println("failed to set prices from binance to db: ", err)
			continue
		}

		time.Sleep(Interval)
	}
}
