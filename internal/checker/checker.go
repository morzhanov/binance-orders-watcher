package checker

import (
	"log"
	"time"

	"github.com/morzhanov/binance-orders-watcher/internal/fetcher"
)

// TODO: checked should get alerts from db and send alert if alert limit reached

type Checker interface {
	Run() error
}

type checkerImpl struct {
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
