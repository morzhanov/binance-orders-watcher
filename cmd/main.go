package main

import (
	"log"

	"github.com/morzhanov/binance-orders-watcher/internal/alertmanager"
	"github.com/morzhanov/binance-orders-watcher/internal/binance"
	"github.com/morzhanov/binance-orders-watcher/internal/checker"
	"github.com/morzhanov/binance-orders-watcher/internal/client"
	"github.com/morzhanov/binance-orders-watcher/internal/config"
	"github.com/morzhanov/binance-orders-watcher/internal/cron"
	"github.com/morzhanov/binance-orders-watcher/internal/db"
	"github.com/morzhanov/binance-orders-watcher/internal/debug"
	"github.com/morzhanov/binance-orders-watcher/internal/fetcher"
)

func main() {
	if debug.IsDebug() {
		log.Println("app started in debug mode: database will not be cleared and cron will not be run")
	}

	conf, err := config.New("./", ".env")
	if err != nil {
		log.Fatal(err)
	}
	alertManager := alertmanager.New(conf.MailjetApiKey, conf.MailjetApiSecret, conf.MailjetSenderName, conf.MailjetSenderEmail)
	dbClient, err := db.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	binClient := binance.New(conf.BinApiKey, conf.BinApiSecret, conf.BinProdURI)
	fetcherClient := fetcher.New(binClient, dbClient)
	checkerClient := checker.New(dbClient, alertManager)

	cronClient := cron.New(fetcherClient, checkerClient)
	cl := client.New(conf.BaseAuthUsername, conf.BaseAuthPassword, conf.BaseAuthSecret, conf.AppURI, conf.AppSchema, conf.AppPort, conf.MailjetSenderName, conf.MailjetSenderEmail, dbClient, fetcherClient, checkerClient, alertManager)

	go func() {
		if debug.IsDebug() {
			log.Println("debug mode, skipping cron start...")
			return
		}
		log.Println("starting cron...")
		if err = cronClient.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	if err = cl.Run(conf.AppTlsCertPath, conf.AppTlsKeyPath); err != nil {
		log.Fatal(err)
	}
}
