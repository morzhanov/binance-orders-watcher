package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	apiKeyHeaderName = "X-MBX-APIKEY"
)

func main() {
	apiKey, ok := os.LookupEnv("BINANCE_API_KEY")
	if !ok {
		log.Fatalf("BINANCE_API_KEY env not found")
	}
	secret, ok := os.LookupEnv("BINANCE_API_SECRET")
	if !ok {
		log.Fatalf("BINANCE_API_SECRET env not found")
	}
	prodURI, ok := os.LookupEnv("BINANCE_PRODUCTION_URI")
	if !ok {
		log.Fatalf("BINANCE_PRODUCTION_URI env not found")
	}

	ts := time.Now()
	query := fmt.Sprintf("timestamp=%d&recvWindow=10000", ts.UnixMilli())

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(query))
	sha := hex.EncodeToString(h.Sum(nil))
	log.Println("sha = ", sha)

	uri := fmt.Sprintf("%s/api/v3/openOrders?%s&signature=%s", prodURI, query, sha)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	req.Header[apiKeyHeaderName] = []string{apiKey}
	res, err := http.DefaultClient.Do(req)
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	log.Println("body = ", bodyString)
}
