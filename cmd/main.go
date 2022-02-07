package main

import (
	"log"
	"os"

	"github.com/morzhanov/binance-orders-watcher/internal/alertmanager"
	"github.com/morzhanov/binance-orders-watcher/internal/binance"
	"github.com/morzhanov/binance-orders-watcher/internal/checker"
	"github.com/morzhanov/binance-orders-watcher/internal/client"
	"github.com/morzhanov/binance-orders-watcher/internal/cron"
	"github.com/morzhanov/binance-orders-watcher/internal/db"
	"github.com/morzhanov/binance-orders-watcher/internal/debug"
	"github.com/morzhanov/binance-orders-watcher/internal/fetcher"
)

func main() {
	if debug.IsDebug() {
		log.Println("app started in debug mode: database will not be cleared and cron will not be run")
	}

	appPort := getEnvVar("APP_PORT")
	appURI := getEnvVar("APP_URI")
	appSchema := getEnvVar("APP_SCHEMA")
	binApiKey := getEnvVar("BINANCE_API_KEY")
	binApiSecret := getEnvVar("BINANCE_API_SECRET")
	BinProdURI := getEnvVar("BINANCE_PRODUCTION_URI")
	baseAuthUsername := getEnvVar("BASE_AUTH_USERNAME")
	baseAuthPassword := getEnvVar("BASE_AUTH_PASSWORD")
	baseAuthSecret := getEnvVar("BASE_AUTH_SECRET")
	mailjetApiKey := getEnvVar("MAILJET_API_KEY")
	mailjetApiSecret := getEnvVar("MAILJET_API_SECRET")
	mailjetSenderName := getEnvVar("MAILJET_SENDER_NAME")
	mailjetSenderEmail := getEnvVar("MAILJET_SENDER_EMAIL")

	alertManager := alertmanager.New(mailjetApiKey, mailjetApiSecret, mailjetSenderName, mailjetSenderEmail)
	dbClient, err := db.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	binClient := binance.New(binApiKey, binApiSecret, BinProdURI)
	fetcherClient := fetcher.New(binClient, dbClient)
	checkerClient := checker.New(dbClient, alertManager)

	cronClient := cron.New(fetcherClient, checkerClient)
	cl := client.New(baseAuthUsername, baseAuthPassword, baseAuthSecret, appURI, appSchema, appPort, dbClient, fetcherClient, checkerClient, alertManager)

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
	if err = cl.Run(); err != nil {
		log.Fatal(err)
	}
}

func getEnvVar(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("%s env variable is not found", key)
	}
	return val
}
