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
	OrderId                int    `json:"orderId"`
	OrderListId            int    `json:"orderListId"`
	ClientOrderId          string `json:"clientOrderId"`
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
	MarketPrice            int    `json:"marketPrice"`
	OrderMarketPriceSpread int    `json:"orderMarketPriceSpread"`
}

type Price struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type Alert struct {
	ID            string `json:"id"`
	Symbol        string `json:"symbol"`
	Price         int    `json:"price"`
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
	insertSQL := `BEGIN TRANSACTION;`

	for _, o := range orders {
		insertSQL += fmt.Sprintf(`
			INSERT INTO orders ('symbol', 'orderId', 'orderListId', 'clientOrderId', 'price', 'origQty', 'executedQty', 'cummulativeQuoteQty', 'status', 'timeInForce', 'type', 'side', 'stopPrice', 'icebergQty', 'time', 'updateTime', 'isWorking', 'marketPrice', 'orderMarketPriceSpread')
			VALUES('%s', '%d', '%d', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%d', '%d', '%t', '%d', '%d');
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
			VALUES('%s', '%s');
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

func (c *client) AddAlert(alert *Alert) error {
	log.Println("inserting alert into db...")
	insertSQL := fmt.Sprintf(`
			INSERT INTO alerts ('id' ,'symbol', 'price', 'name', 'email', 'text', 'directionDown')
			VALUES('%s', '%s', '%d', '%s', '%s', '%s', '%t');
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
		"marketPrice" INTEGER,
		"orderMarketPriceSpread" INTEGER
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
		"price" INTEGER,
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
