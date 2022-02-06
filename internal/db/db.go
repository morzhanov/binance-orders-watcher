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
	GetOrders() ([]*Order, error)
	SetPrices(prices []*Price) error
	GetPrices() ([]*Price, error)
	AddAlert(alert *Alert) error
	DeleteAlert(alert *Alert) error
	GetAlerts() ([]*Alert, error)
}

type client struct {
	db *sql.DB
}

type Order struct {
	Symbol                 string `json:"symbol"`
	OrderID                int    `json:"orderId"`
	OrderListID            int    `json:"orderListId"`
	ClientOrderID          string `json:"clientOrderId"`
	Price                  string `json:"price"`
	OrigQty                string `json:"origQty"`
	ExecutedQty            string `json:"executedQty"`
	CummulativeQuoteQty    string `json:"cummulativeQuoteQty"`
	Status                 string `json:"status"`
	TimeInForce            string `json:"timeInForce"`
	Type                   string `json:"type"`
	Side                   string `json:"side"`
	StopPrice              string `json:"stopPrice"`
	IcebergQty             string `json:"icebergQty"`
	Time                   int    `json:"time"`
	UpdateTime             int    `json:"updateTime"`
	IsWorking              bool   `json:"isWorking"`
	LastOrderPrice         string `json:"lastOrderPrice"`
	MarketPrice            string `json:"marketPrice"`
	PercentCompleted       string `json:"percentCompleted"`
	OrderMarketPriceSpread string `json:"orderMarketPriceSpread"`
}

type Price struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type Alert struct {
	ID            string `json:"id"`
	Symbol        string `json:"symbol"`
	Price         string `json:"price"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Text          string `json:"text"`
	DirectionDown bool   `json:"directionDown"`
}

func NewClient() (Client, error) {
	if instance != nil {
		return instance, nil
	}

	if err := os.Remove("sqlite-database.db"); err != nil {
		log.Println("sqlite-database.db is not exists")
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
	if err := c.deleteOrders(); err != nil {
		return err
	}
	for _, o := range orders {
		if err := c.createOrder(o); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) createOrder(o *Order) error {
	insertSQL := fmt.Sprintf(`
			INSERT INTO orders ('symbol', 'orderId', 'orderListId', 'clientOrderId', 'price', 'origQty', 'executedQty', 'cummulativeQuoteQty', 'status', 'timeInForce', 'type', 'side', 'stopPrice', 'icebergQty', 'time', 'updateTime', 'isWorking', 'lastOrderPrice', 'marketPrice', 'percentCompleted', 'orderMarketPriceSpread')
			VALUES('%s', '%d', '%d', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%d', '%d', '%t', '%s', '%s', '%s', '%s');
	`, o.Symbol, o.OrderID, o.OrderListID, o.ClientOrderID, o.Price, o.OrigQty, o.ExecutedQty, o.CummulativeQuoteQty, o.Status, o.TimeInForce, o.Type, o.Side, o.StopPrice, o.IcebergQty, o.Time, o.UpdateTime, o.IsWorking, o.LastOrderPrice, o.MarketPrice, o.PercentCompleted, o.OrderMarketPriceSpread)
	statement, err := c.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

func (c *client) deleteOrders() error {
	deleteSQL := fmt.Sprintf("DELETE FROM orders")
	statement, err := c.db.Prepare(deleteSQL)
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
		err = row.Scan(&order.Symbol, &order.OrderID, &order.OrderListID, &order.ClientOrderID, &order.Price, &order.OrigQty, &order.ExecutedQty, &order.CummulativeQuoteQty, &order.Status, &order.TimeInForce, &order.Type, &order.Side, &order.StopPrice, &order.IcebergQty, &order.Time, &order.UpdateTime, &order.IsWorking, &order.LastOrderPrice, &order.MarketPrice, &order.PercentCompleted, &order.OrderMarketPriceSpread)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (c *client) SetPrices(prices []*Price) error {
	log.Println("inserting price records into db...")
	if err := c.deletePrices(); err != nil {
		return err
	}
	for _, p := range prices {
		if err := c.createPrice(p); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) createPrice(p *Price) error {
	insertSQL := fmt.Sprintf(`
			INSERT INTO prices ('symbol', 'price')
			VALUES('%s', '%s');
		`, p.Symbol, p.Price)
	statement, err := c.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

func (c *client) deletePrices() error {
	deleteSQL := fmt.Sprintf("DELETE FROM prices")
	statement, err := c.db.Prepare(deleteSQL)
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

func (c *client) AddAlert(alert *Alert) error {
	log.Println("inserting alert into db...")
	insertSQL := fmt.Sprintf(`
			INSERT INTO alerts ('id' ,'symbol', 'price', 'name', 'email', 'text', 'directionDown')
			VALUES('%s', '%s', '%s', '%s', '%s', '%s', '%t');
	`, alert.ID, alert.Symbol, alert.Price, alert.Name, alert.Email, alert.Text, alert.DirectionDown)

	statement, err := c.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

func (c *client) DeleteAlert(alert *Alert) error {
	log.Printf("deleting alert with id %s...", alert.ID)
	deleteSQL := fmt.Sprintf(`
		DELETE FROM alerts
		WHERE id = '%s'
	`, alert.ID)

	statement, err := c.db.Prepare(deleteSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

func (c *client) GetAlerts() ([]*Alert, error) {
	log.Println("getting alert records from db...")
	row, err := c.db.Query("SELECT * FROM alerts")
	if err != nil {
		return nil, err
	}
	defer row.Close()

	alerts := make([]*Alert, 0)
	for row.Next() {
		alert := &Alert{}
		err = row.Scan(&alert.ID, &alert.Symbol, &alert.Price, &alert.Name, &alert.Email, &alert.Text, &alert.DirectionDown)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, nil
}

func (c *client) createTables() error {
	ordersTableSQL := `CREATE TABLE orders (		
		"symbol" TEXT,
		"orderId" INTEGER,
		"orderListId" INTEGER,
		"clientOrderId" TEXT,
		"price" TEXT,
		"origQty" TEXT,
		"executedQty" TEXT,
		"cummulativeQuoteQty" TEXT,
		"status" TEXT,
		"timeInForce" TEXT,
		"type" TEXT,
		"side" TEXT,
		"stopPrice" TEXT,
		"icebergQty" TEXT,
		"time" INTEGER,
		"updateTime" INTEGER,
		"isWorking" BOOLEAN,
		"lastOrderPrice" TEXT,
		"marketPrice" TEXT,
		"percentCompleted" TEXT,
		"orderMarketPriceSpread" TEXT
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
		"symbol" TEXT,
		"price" TEXT	
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

	alertsTableSQL := `CREATE TABLE alerts (		
		"id" TEXT,
		"symbol" TEXT,
		"price" TEXT,
		"name" TEXT,
		"email" TEXT,
		"text" TEXT,
		"directionDown" BOOLEAN
	  );`

	log.Println("create alerts table...")
	statement, err = c.db.Prepare(alertsTableSQL)
	if err != nil {
		return err
	}
	if _, err = statement.Exec(); err != nil {
		return err
	}
	log.Println("alerts table created")
	return nil
}
