package client

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/morzhanov/binance-orders-watcher/internal/db"

	"github.com/form3tech-oss/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/morzhanov/binance-orders-watcher/internal/alertmanager"
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
	authUsername string
	authPassword string
	authSecret   string
	db           db.Client
	alertManager alertmanager.Manager
}

type JWTPayload struct {
	ID        string    `json:"id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (payload *JWTPayload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return errors.New("token is expired")
	}
	return nil
}

func New(authUsername, authPassword, authSecret string, db db.Client, alertManager alertmanager.Manager) Client {
	c := &client{
		authUsername: authUsername,
		authPassword: authPassword,
		authSecret:   authSecret,
		db:           db,
		alertManager: alertManager,
	}

	r := mux.NewRouter()
	r.Use(c.authMiddleware)
	r.HandleFunc("/", c.homeHandler)
	http.Handle("/", r)
	c.r = r

	return c
}

func (c *client) Run(port string) error {
	return http.ListenAndServe(":"+port, c.r)
}

func (c *client) homeHandler(w http.ResponseWriter, r *http.Request) {
	orders := c.db.GetOrders()
	prices := c.db.GetPrices()

	// TODO: create template and render orders/prices, alerts and gaps
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
