package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/morzhanov/binance-orders-watcher/internal/debug"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

const (
	dbFileName = "sqlite-database.db"
)

var instance Client

type Client interface {
	SetOrders(orders []*Order) error
	GetOrders() ([]*Order, error)
	SetPrices(prices []*Price) error
	GetPrices() ([]*Price, error)
	AddAlert(alert *Alert) error
	DeleteAlert(id string) error
	GetAlerts() ([]*Alert, error)
	AddAuthRequest(ip string) error
	UpdateAuthRequest(ip string, attempts int, alertSent bool) error
	GetAuthRequest(ip string) (*AuthRequest, error)
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

type AuthRequest struct {
	IP        string `json:"ip"`
	Attempts  int    `json:"attempts"`
	AlertSent bool   `json:"alertSent"`
}

func NewClient() (Client, error) {
	if instance != nil {
		return instance, nil
	}

	var sqlDB *sql.DB
	if !debug.IsDebug() && !dbExists() {
		if err := os.Remove(dbFileName); err != nil {
			log.Println(dbFileName + " is not exists")
		}
		log.Println("creating " + dbFileName)
		file, err := os.Create(dbFileName)
		if err != nil {
			return nil, err
		}
		if err = file.Close(); err != nil {
			return nil, err
		}

		sqlDB, _ = sql.Open("sqlite3", "./"+dbFileName)
		if err = createTables(sqlDB); err != nil {
			return nil, err
		}
	} else {
		sqlDB, _ = sql.Open("sqlite3", "./"+dbFileName)
	}

	return &client{db: sqlDB}, nil
}

func dbExists() bool {
	if _, err := os.Stat("./" + dbFileName); err != nil {
		return false
	}
	return true
}

func createTables(sqlDB *sql.DB) error {
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
	statement, err := sqlDB.Prepare(ordersTableSQL)
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
	statement, err = sqlDB.Prepare(pricesTableSQL)
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
	statement, err = sqlDB.Prepare(alertsTableSQL)
	if err != nil {
		return err
	}
	if _, err = statement.Exec(); err != nil {
		return err
	}
	log.Println("alerts table created")

	authRequestsTableSQL := `CREATE TABLE auth_requests (		
		"ip" TEXT,
		"attempts" INTEGER,
		"alertSent" BOOLEAN
	  );`

	log.Println("create auth requests table...")
	statement, err = sqlDB.Prepare(authRequestsTableSQL)
	if err != nil {
		return err
	}
	if _, err = statement.Exec(); err != nil {
		return err
	}
	log.Println("auth requests table created")
	return nil
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

func (c *client) DeleteAlert(id string) error {
	log.Printf("deleting alert with id %s...", id)
	deleteSQL := fmt.Sprintf(`
		DELETE FROM alerts
		WHERE id = '%s'
	`, id)

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

func (c *client) AddAuthRequest(ip string) error {
	log.Println("inserting auth request into db...")
	insertSQL := fmt.Sprintf(`
			INSERT INTO auth_requests ('ip' ,'attempts', 'alertSent')
			VALUES('%s', '%d', '%t');
	`, ip, 0, false)

	statement, err := c.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

func (c *client) UpdateAuthRequest(ip string, attempts int, alertSent bool) error {
	updateSQL := fmt.Sprintf(`
			UPDATE auth_requests
			SET attempts = '%d', 'alertSent' = %t
			WHERE ip = '%s';
	`, attempts, alertSent, ip)

	statement, err := c.db.Prepare(updateSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

func (c *client) GetAuthRequest(ip string) (*AuthRequest, error) {
	getSQL := fmt.Sprintf(`
		SELECT * FROM auth_requests
		WHERE ip = '%s';`,
		ip)

	row := c.db.QueryRow(getSQL)
	req := &AuthRequest{}
	err := row.Scan(&req.IP, &req.Attempts, &req.AlertSent)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, err
	}
	return req, nil
}
