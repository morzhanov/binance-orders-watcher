package cron

import (
	"log"
	"time"

	"github.com/morzhanov/binance-orders-watcher/internal/fetcher"
)

const (
	Interval = time.Minute * 30
)

type Cron interface {
	Run() error
}

type cronImp struct {
	fetcher fetcher.Fetcher
}

func New(fetcherClient fetcher.Fetcher) Cron {
	return &cronImp{fetcher: fetcherClient}
}

func (c *cronImp) Run() error {
	for {
		if err := c.fetcher.Fetch(); err != nil {
			log.Println("error in fetcher: ", err)
		}
		time.Sleep(Interval)
	}
}
