package client

import (
	"crypto/subtle"
	"errors"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/morzhanov/binance-orders-watcher/internal/alertmanager"
	"github.com/morzhanov/binance-orders-watcher/internal/db"
	"github.com/morzhanov/binance-orders-watcher/internal/fetcher"
)

const (
	AuthCookieName               = "access_token"
	BearerTokenPrefix            = "Bearer "
	TokenExpirationDurationInSec = 84600
)

type Client interface {
	Run(port string) error
}

type client struct {
	r            *mux.Router
	appUri       string
	authUsername string
	authPassword string
	authSecret   string
	db           db.Client
	fetcher      fetcher.Fetcher
	alertManager alertmanager.Manager
}

type JWTPayload struct {
	ID        string    `json:"id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

type HomePageTemplateData struct {
	AppURI string
	Orders []*db.Order
	Prices []*db.Price
}

func (payload *JWTPayload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return errors.New("token is expired")
	}
	return nil
}

func New(authUsername, authPassword, authSecret, appUri string, dbClient db.Client, fetcherClient fetcher.Fetcher, alertManager alertmanager.Manager) Client {
	c := &client{
		appUri:       appUri,
		authUsername: authUsername,
		authPassword: authPassword,
		authSecret:   authSecret,
		db:           dbClient,
		fetcher:      fetcherClient,
		alertManager: alertManager,
	}

	// TODO: we should add an endpoint to configure limit when alert should be sent
	r := mux.NewRouter()
	r.Use(c.authMiddleware)
	r.HandleFunc("/", c.homeHandler)
	r.HandleFunc("/refresh", c.refreshDataHandler)
	c.r = r

	return c
}

func (c *client) Run(port string) error {
	return http.ListenAndServe(":"+port, c.r)
}

func (c *client) homeHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.ParseFiles("./templates/home.html")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	orders := c.db.GetOrders()
	prices := c.db.GetPrices()
	homePageData := &HomePageTemplateData{
		AppURI: c.appUri,
		Orders: orders,
		Prices: prices,
	}
	if err = tmpl.Execute(w, homePageData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func (c *client) refreshDataHandler(w http.ResponseWriter, _ *http.Request) {
	if err := c.fetcher.Fetch(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data successfully loaded from Binance, the page could be reloaded."))
	return
}

func (c *client) authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.checkAccessToken(w, r) {
			h.ServeHTTP(w, r)
			return
		}
		token := c.basicAuth(w, r)
		if token == "" {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
		authCookie := &http.Cookie{
			Name:    AuthCookieName,
			Value:   BearerTokenPrefix + token,
			Expires: time.Now().Add(time.Second * TokenExpirationDurationInSec),
		}
		http.SetCookie(w, authCookie)
	})
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
		return nil, nil
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
		return ""
	}

	token, err := c.createAccessToken()
	if err != nil {
		return ""
	}
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
