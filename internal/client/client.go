package client

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/morzhanov/binance-orders-watcher/internal/alertmanager"

	"github.com/morzhanov/binance-orders-watcher/internal/checker"

	"github.com/form3tech-oss/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/morzhanov/binance-orders-watcher/internal/db"
	"github.com/morzhanov/binance-orders-watcher/internal/fetcher"
)

const (
	AuthCookieName               = "access_token"
	BearerTokenPrefix            = "Bearer "
	TokenExpirationDurationInSec = 84600
	AppSchemaHTTPS               = "https"
)

type Client interface {
	Run(tlsCertPath, tlsKeyPath string) error
}

type client struct {
	r                      *mux.Router
	appSchema              string
	appPort                string
	appUri                 string
	authUsername           string
	authPassword           string
	authSecret             string
	authReqAlertAdminName  string
	authReqAlertAdminEmail string
	db                     db.Client
	fetcher                fetcher.Fetcher
	checker                checker.Checker
	alertManager           alertmanager.Manager
}

type JWTPayload struct {
	ID        string    `json:"id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

type HomePageTemplateData struct {
	AppURI    string
	AppSchema string
	AppPort   string
	Orders    []*db.Order
	Prices    []*db.Price
	Alerts    []*db.Alert
}

func (payload *JWTPayload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return errors.New("token is expired")
	}
	return nil
}

func New(authUsername, authPassword, authSecret, appUri, appSchema, appPort, authReqAlertAdminName, authReqAlertAdminEmail string, dbClient db.Client, fetcherClient fetcher.Fetcher, checker checker.Checker, alertManager alertmanager.Manager) Client {
	c := &client{
		appUri:                 appUri,
		appSchema:              appSchema,
		appPort:                appPort,
		authUsername:           authUsername,
		authPassword:           authPassword,
		authSecret:             authSecret,
		authReqAlertAdminName:  authReqAlertAdminName,
		authReqAlertAdminEmail: authReqAlertAdminEmail,
		db:                     dbClient,
		fetcher:                fetcherClient,
		checker:                checker,
		alertManager:           alertManager,
	}

	r := mux.NewRouter()
	r.Use(c.authMiddleware)
	r.HandleFunc("/", c.homeHandler)
	r.HandleFunc("/refresh", c.refreshDataHandler)
	r.HandleFunc("/alert", c.addAlertHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./internal/client/static/")))
	c.r = r

	return c
}

func (c *client) Run(tlsCertPath, tlsKeyPath string) error {
	log.Printf("starting client application on %s://%s:%s", c.appSchema, c.appUri, c.appPort)
	addr := ":" + c.appPort
	if c.appSchema == AppSchemaHTTPS && tlsCertPath != "" && tlsKeyPath != "" {
		return http.ListenAndServeTLS(addr, tlsCertPath, tlsKeyPath, c.r)
	}
	return http.ListenAndServe(addr, c.r)
}

func (c *client) homeHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.ParseFiles("./internal/client/templates/home.html")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	orders, err := c.db.GetOrders()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	prices, err := c.db.GetPrices()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	alerts, err := c.db.GetAlerts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	homePageData := &HomePageTemplateData{
		AppURI:    c.appUri,
		AppSchema: c.appSchema,
		AppPort:   c.appPort,
		Orders:    orders,
		Prices:    prices,
		Alerts:    alerts,
	}
	if err = tmpl.Execute(w, homePageData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func (c *client) refreshDataHandler(w http.ResponseWriter, _ *http.Request) {
	_, prices, err := c.fetcher.Fetch()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if err = c.checker.Check(prices); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data successfully loaded from Binance, the page could be reloaded."))
	return
}

func (c *client) addAlertHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	log.Println("unmarshalling alert object: ", string(body))
	var alert db.Alert
	if err = json.Unmarshal(body, &alert); err != nil {
		log.Println("failed to unmarshal alert object: ", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	alertID, err := uuid.NewUUID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	alert.ID = alertID.String()
	err = c.db.AddAlert(&alert)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alert successfully created"))
	return
}

func (c *client) authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.checkAccessToken(w, r) {
			h.ServeHTTP(w, r)
			return
		}

		if !c.checkAuthAttempts(r) {
			w.WriteHeader(403)
			w.Write([]byte("Forbidden.\n"))
			return
		}

		log.Println("access token is not found in cookies, starting base auth flow...")
		token := c.basicAuth(w, r)
		log.Println("performed basic auth, token = ", token)
		if token == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		if err := c.clearAuthAttempts(r); err != nil {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
		}

		log.Println("creating auth cookie")
		authCookie := &http.Cookie{
			Name:    AuthCookieName,
			Value:   BearerTokenPrefix + token,
			Expires: time.Now().Add(time.Second * TokenExpirationDurationInSec),
		}
		http.SetCookie(w, authCookie)
		h.ServeHTTP(w, r)
	})
}

func (c *client) checkAuthAttempts(r *http.Request) bool {
	ip := readUserIP(r)
	req, err := c.db.GetAuthRequest(ip)
	if err != nil {
		return false
	}

	if req == nil {
		if err = c.db.AddAuthRequest(ip); err != nil {
			return false
		}
		return true
	}
	if req.Attempts >= 3 {
		if !req.AlertSent {
			text := fmt.Sprintf("User with IP %s is blocket: auth req count threshold reached.", ip)
			c.alertManager.SendAlert(c.authReqAlertAdminEmail, c.authReqAlertAdminName, text)
			c.db.UpdateAuthRequest(ip, 3, true)
		}
		return false
	}

	if err = c.db.UpdateAuthRequest(ip, req.Attempts+1, false); err != nil {
		return false
	}
	return true
}

func (c *client) clearAuthAttempts(r *http.Request) error {
	ip := readUserIP(r)
	return c.db.UpdateAuthRequest(ip, 0, false)
}

func (c *client) checkAccessToken(_ http.ResponseWriter, r *http.Request) bool {
	authCookie, err := r.Cookie(AuthCookieName)
	if err != nil || authCookie == nil || authCookie.Value == "" {
		return false
	}
	token := strings.Replace(authCookie.Value, BearerTokenPrefix, "", 1)
	return c.verifyToken(token)
}

func (c *client) verifyToken(token string) bool {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("token is not valid")
		}
		return []byte(c.authSecret), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &JWTPayload{}, keyFunc)
	if err != nil {
		return false
	}
	if _, ok := jwtToken.Claims.(*JWTPayload); !ok {
		return false
	}
	return true
}

func (c *client) basicAuth(_ http.ResponseWriter, r *http.Request) string {
	user, pass, ok := r.BasicAuth()
	if !ok ||
		subtle.ConstantTimeCompare([]byte(user), []byte(c.authUsername)) != 1 ||
		subtle.ConstantTimeCompare([]byte(pass), []byte(c.authPassword)) != 1 {
		log.Println("basic auth failed")
		log.Printf("user = %s, pass = %s, c.AuthUser = %s, c.authPass = %s", user, pass, c.authUsername, c.authPassword)
		return ""
	}

	token, err := c.createAccessToken()
	if err != nil {
		log.Println("basic auth failed, ", err)
		return ""
	}
	log.Println("basic auth succeeded, token = ", token)
	return token
}

func (c *client) createAccessToken() (string, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	payload := &JWTPayload{
		ID:        tokenID.String(),
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(time.Second * TokenExpirationDurationInSec),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(c.authSecret))
}

func readUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
