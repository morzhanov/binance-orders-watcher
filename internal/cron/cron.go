package cron

import (
	"log"
	"time"

	"github.com/morzhanov/binance-orders-watcher/internal/checker"

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
	checker checker.Checker
}

func New(fetcherClient fetcher.Fetcher, checkerClient checker.Checker) Cron {
	return &cronImp{fetcher: fetcherClient, checker: checkerClient}
}

func (c *cronImp) Run() error {
	for {
		_, prices, err := c.fetcher.Fetch()
		if err != nil {
			log.Println("error in fetcher: ", err)
		}
		if err = c.checker.Check(prices); err != nil {
			return err
		}
		time.Sleep(Interval)
	}
}
