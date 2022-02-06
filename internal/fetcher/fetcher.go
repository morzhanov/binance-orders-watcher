package fetcher

import (
	"fmt"
	"log"
	"strconv"

	"github.com/morzhanov/binance-orders-watcher/internal/binance"
	"github.com/morzhanov/binance-orders-watcher/internal/db"
)

const (
	orderStatusFilled = "FILLED"
	notAvailableText  = "N/A"
)

type Fetcher interface {
	Fetch() (orders []*db.Order, prices []*db.Price, err error)
}

type fetcherImp struct {
	binClient binance.Client
	db        db.Client
}

func New(binClient binance.Client, dbClient db.Client) Fetcher {
	return &fetcherImp{binClient: binClient, db: dbClient}
}

func (f *fetcherImp) Fetch() ([]*db.Order, []*db.Price, error) {
	binanceOrders, err := f.binClient.GetOrders()
	if err != nil {
		log.Println("failed to get binanceOrders from binance: ", err)
		return nil, nil, err
	}
	prices, err := f.binClient.GetPrices()
	if err != nil {
		log.Println("failed to get prices from binance: ", err)
		return nil, nil, err
	}

	orders, err := f.binanceOrdersToDBOrders(binanceOrders, prices)
	if err != nil {
		return nil, nil, err
	}
	if err = f.db.SetOrders(orders); err != nil {
		log.Println("failed to set binanceOrders from binance to db: ", err)
		return nil, nil, err
	}
	if err = f.db.SetPrices(prices); err != nil {
		log.Println("failed to set prices from binance to db: ", err)
		return nil, nil, err
	}
	return orders, prices, nil
}

func (f *fetcherImp) binanceOrdersToDBOrders(binOrders []*binance.BinanceOrder, prices []*db.Price) ([]*db.Order, error) {
	allOrders := make(map[string][]*binance.BinanceOrder, 0)
	var orders []*db.Order
	var err error

	for _, binOrder := range binOrders {
		var parsedMarketPrice, parsedOrderPrice float64
		var marketPrice, spread string
		for _, price := range prices {
			if price.Symbol == binOrder.Symbol {
				parsedMarketPrice, err = strconv.ParseFloat(price.Price, 64)
				if err != nil {
					return nil, err
				}
				parsedOrderPrice, err = strconv.ParseFloat(binOrder.Price, 64)
				if err != nil {
					return nil, err
				}
				marketPrice = price.Price
				spreadVal := parsedOrderPrice - parsedMarketPrice
				spread = fmt.Sprintf("%f", spreadVal)
				break
			}
		}

		allOrdersForSymbol, ok := allOrders[binOrder.Symbol]
		if !ok {
			res, err := f.binClient.GetAllOrdersForSymbol(binOrder.Symbol)
			if err != nil {
				return nil, err
			}
			allOrdersForSymbol = res
		}
		allOrders[binOrder.Symbol] = allOrdersForSymbol

		var lastOrderPrice string
		for _, order := range allOrdersForSymbol {
			if order.Status == orderStatusFilled {
				lastOrderPrice = order.Price
				break
			}
		}
		if lastOrderPrice == "" {
			lastOrderPrice = notAvailableText
		}

		var percentCompleted string
		if lastOrderPrice == notAvailableText {
			percentCompleted = notAvailableText
		} else {
			originalPrice, err := strconv.ParseFloat(lastOrderPrice, 64)
			if err != nil {
				return nil, err
			}
			percentCompleted = calculateOrderPercentCompleted(originalPrice, parsedMarketPrice, parsedOrderPrice)
		}

		order := &db.Order{
			Symbol:                 binOrder.Symbol,
			OrderID:                binOrder.OrderId,
			OrderListID:            binOrder.OrderListId,
			ClientOrderID:          binOrder.ClientOrderId,
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
			LastOrderPrice:         lastOrderPrice,
			MarketPrice:            marketPrice,
			PercentCompleted:       percentCompleted,
			OrderMarketPriceSpread: spread,
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func calculateOrderPercentCompleted(original, market, price float64) string {
	var res float64
	if price > original {
		res = ((market - original) * 100) / (price - original)
	} else {
		res = ((original - market) * 100) / (original - price)
	}
	return fmt.Sprintf("%d", int(res))
}
