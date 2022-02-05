package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

var instance Client

type Client interface {
	SetOrders(orders []*Order) error
	GetOrders() []*Order
	SetPrices(prices []*Price) error
	GetPrices() []*Price
}

type client struct {
	db *sql.DB
}

// TODO: we should store limits and alerts in the database
type Order struct {
	Symbol                 string `json:"symbol"`
	OrderId                int    `json:"orderId"`
	OrderListId            int    `json:"orderListId"`
	ClientOrderId          string `json:"clientOrderId"`
	Price                  int    `json:"price"`
	OrigQty                int    `json:"origQty"`
	ExecutedQty            int    `json:"executedQty"`
	CummulativeQuoteQty    int    `json:"cummulativeQuoteQty"`
	Status                 string `json:"status"`
	TimeInForce            string `json:"timeInForce"`
	Type                   string `json:"type"`
	Side                   string `json:"side"`
	StopPrice              int    `json:"stopPrice"`
	IcebergQty             int    `json:"icebergQty"`
	Time                   string `json:"time"`
	UpdateTime             string `json:"updateTime"`
	IsWorking              bool   `json:"isWorking"`
	MarketPrice            int    `json:"marketPrice"`
	OrderMarketPriceSpread int    `json:"orderMarketPriceSpread"`
}

type Price struct {
	Symbol string `json:"symbol"`
	Price  int    `json:"price"`
}

func NewClient() (Client, error) {
	if instance != nil {
		return instance, nil
	}

	if err := os.Remove("sqlite-database.db"); err != nil {
		return nil, err
	}
	log.Println("creating sqlite-database.db...")
	file, err := os.Create("sqlite-database.db") // Create SQLite file
	if err != nil {
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	log.Println("sqlite-database.db created")
	sqlDB, _ := sql.Open("sqlite3", "./sqlite-database.db")

	c := &client{db: sqlDB}
	if err = c.createTables(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *client) SetOrders(orders []*Order) error {
	log.Println("inserting order records into db...")
	insertSQL := `BEGIN TRANSACTION;`

	for _, o := range orders {
		insertSQL += fmt.Sprintf(`
			INSERT INTO orders ('symbol', 'orderId', 'orderListId', 'clientOrderId', 'price', 'origQty', 'executedQty', 'cummulativeQuoteQty', 'status', 'timeInForce', 'type', 'side', 'stopPrice', 'icebergQty', 'time', 'updateTime', 'isWorking', 'marketPrice', 'orderMarketPriceSpread')
			VALUES('%s', '%d', '%d', '%s', '%d', '%d', '%d', '%d', '%s', '%s', '%s', '%s', '%d', '%d', '%s', '%s', '%t', '%d', '%d');
		`, o.Symbol, o.OrderId, o.OrderListId, o.ClientOrderId, o.Price, o.OrigQty, o.ExecutedQty, o.CummulativeQuoteQty, o.Status, o.TimeInForce, o.Type, o.Side, o.StopPrice, o.IcebergQty, o.Time, o.UpdateTime, o.IsWorking, o.MarketPrice, o.OrderMarketPriceSpread)
	}
	insertSQL += `COMMIT;`

	statement, err := c.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

func (c *client) GetOrders() ([]*Order, error) {
	log.Println("getting order records from db...")
	row, err := c.db.Query("SELECT * FROM orders")
	if err != nil {
		return nil, err
	}
	defer row.Close()

	orders := make([]*Order, 0)
	for row.Next() {
		order := &Order{}
		err = row.Scan(&order.Symbol, &order.OrderId, &order.OrderListId, &order.ClientOrderId, &order.Price, &order.OrigQty, &order.ExecutedQty, &order.CummulativeQuoteQty, &order.Status, &order.TimeInForce, &order.Type, &order.Side, &order.StopPrice, &order.IcebergQty, &order.Time, &order.UpdateTime, &order.IsWorking, &order.MarketPrice, &order.OrderMarketPriceSpread)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (c *client) SetPrices(orders []*Price) error {
	log.Println("inserting price records into db...")
	insertSQL := `BEGIN TRANSACTION;`

	for _, o := range orders {
		insertSQL += fmt.Sprintf(`
			INSERT INTO prices ('symbol', 'price')
			VALUES('%s', '%d');
		`, o.Symbol, o.Price)
	}
	insertSQL += `COMMIT;`

	statement, err := c.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

func (c *client) GetPrices() ([]*Price, error) {
	log.Println("getting price records from db...")
	row, err := c.db.Query("SELECT * FROM prices")
	if err != nil {
		return nil, err
	}
	defer row.Close()

	orders := make([]*Price, 0)
	for row.Next() {
		order := &Price{}
		err = row.Scan(&order.Symbol, &order.Price)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (c *client) createTables() error {
	ordersTableSQL := `CREATE TABLE orders (		
		"code" TEXT,
		"name" TEXT,
		"program" TEXT		
	  );`

	log.Println("create orders table...")
	statement, err := c.db.Prepare(ordersTableSQL)
	if err != nil {
		return err
	}
	if _, err = statement.Exec(); err != nil {
		return err
	}
	log.Println("orders table created")

	pricesTableSQL := `CREATE TABLE prices (		
		"price" INTEGER,
		"symbol" TEXT		
	  );`

	log.Println("create prices table...")
	statement, err = c.db.Prepare(pricesTableSQL)
	if err != nil {
		return err
	}
	if _, err = statement.Exec(); err != nil {
		return err
	}
	log.Println("prices table created")
	return nil
}
