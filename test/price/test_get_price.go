package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	apiKeyHeaderName = "X-MBX-APIKEY"
)

func main() {
	apiKey, ok := os.LookupEnv("BINANCE_API_KEY")
	if !ok {
		log.Fatalf("BINANCE_API_KEY env not found")
	}
	prodURI, ok := os.LookupEnv("BINANCE_PRODUCTION_URI")
	if !ok {
		log.Fatalf("BINANCE_PRODUCTION_URI env not found")
	}

	uri := fmt.Sprintf("%s/api/v3/ticker/price", prodURI)
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
