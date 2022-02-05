package fetcher

import (
	"log"

	"github.com/morzhanov/binance-orders-watcher/internal/binance"
	"github.com/morzhanov/binance-orders-watcher/internal/db"
)

type Fetcher interface {
	Fetch() error
}

type fetcherImp struct {
	binClient binance.Client
	db        db.Client
}

func New(binClient binance.Client, dbClient db.Client) Fetcher {
	return &fetcherImp{binClient: binClient, db: dbClient}
}

func (f *fetcherImp) Fetch() error {
	binanceOrders, err := f.binClient.GetOrders()
	if err != nil {
		log.Println("failed to get binanceOrders from binance: ", err)
		return err
	}
	prices, err := f.binClient.GetPrices()
	if err != nil {
		log.Println("failed to get prices from binance: ", err)
		return err
	}

	orders := f.binanceOrdersToDBOrders(binanceOrders, prices)
	if err = f.db.SetOrders(orders); err != nil {
		log.Println("failed to set binanceOrders from binance to db: ", err)
		return err
	}
	if err = f.db.SetPrices(prices); err != nil {
		log.Println("failed to set prices from binance to db: ", err)
		return err
	}
	return nil
}

func (f *fetcherImp) binanceOrdersToDBOrders(binOrders []*binance.BinanceOrder, prices []*db.Price) []*db.Order {
	var orders []*db.Order
	for _, binOrder := range binOrders {
		var marketPrice, spread int
		for _, price := range prices {
			if price.Symbol == binOrder.Symbol {
				marketPrice = price.Price
				spread = binOrder.Price - marketPrice
				break
			}
		}

		order := &db.Order{
			Symbol:                 binOrder.Symbol,
			OrderId:                binOrder.OrderId,
			OrderListId:            binOrder.OrderListId,
			ClientOrderId:          binOrder.ClientOrderId,
			Price:                  binOrder.Price,
			OrigQty:                binOrder.OrigQty,
			ExecutedQty:            binOrder.ExecutedQty,
			CummulativeQuoteQty:    binOrder.CummulativeQuoteQty,
			Status:                 binOrder.Status,
			TimeInForce:            binOrder.TimeInForce,
			Type:                   binOrder.Type,
			Side:                   binOrder.Side,
			StopPrice:              binOrder.StopPrice,
			IcebergQty:             binOrder.IcebergQty,
			Time:                   binOrder.Time,
			UpdateTime:             binOrder.UpdateTime,
			IsWorking:              binOrder.IsWorking,
			MarketPrice:            marketPrice,
			OrderMarketPriceSpread: spread,
		}
		orders = append(orders, order)
	}
	return orders
}
