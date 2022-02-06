package checker

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/morzhanov/binance-orders-watcher/internal/alertmanager"
	"github.com/morzhanov/binance-orders-watcher/internal/db"
)

type Checker interface {
	Check(prices []*db.Price) error
}

type checkerImp struct {
	db           db.Client
	alertManager alertmanager.Manager
}

func New(dbClient db.Client, alertManager alertmanager.Manager) Checker {
	return &checkerImp{db: dbClient, alertManager: alertManager}
}

func (c *checkerImp) Check(prices []*db.Price) error {
	log.Println("checking alerts...")
	alerts, err := c.db.GetAlerts()
	if err != nil {
		return err
	}

	for _, alert := range alerts {
		var currentPrice float64
		for _, price := range prices {
			if price.Symbol == alert.Symbol {
				currentPrice, err = strconv.ParseFloat(price.Price, 64)
				if err != nil {
					return err
				}
				break
			}
		}
		if currentPrice == 0 {
			return errors.New(fmt.Sprintf("price for symbol %s is not found in prices array", alert.Symbol))
		}

		parsedAlertPrice, err := strconv.ParseFloat(alert.Price, 64)
		if err != nil {
			return err
		}
		if alert.DirectionDown && parsedAlertPrice <= currentPrice || !alert.DirectionDown && parsedAlertPrice >= currentPrice {
			log.Printf("sending alert for symbol %s: price %d near limit %d", alert.Symbol, alert.Price, currentPrice)
			text := fmt.Sprintf("Binance Order ALERT! Order %s price %d near limit %d", alert.Symbol, alert.Price, currentPrice)
			if err = c.alertManager.SendAlert(alert.Email, alert.Name, text); err != nil {
				return err
			}
			if err = c.db.DeleteAlert(alert); err != nil {
				return err
			}
		}
	}
	return nil
}
