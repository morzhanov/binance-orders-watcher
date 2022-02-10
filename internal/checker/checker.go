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
		if alert.DirectionDown && currentPrice <= parsedAlertPrice || !alert.DirectionDown && currentPrice >= parsedAlertPrice {
			log.Printf("sending alert for symbol %s: price %s near limit %f", alert.Symbol, alert.Price, currentPrice)
			text := fmt.Sprintf("Binance Order ALERT! Order %s price %s near limit %f", alert.Symbol, alert.Price, currentPrice)
			if alert.Text != "" {
				text += "\n\n Additional info: " + alert.Text
			}
			if err = c.alertManager.SendAlert(alert.Email, alert.Name, text); err != nil {
				return err
			}
			if err = c.db.DeleteAlert(alert.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
