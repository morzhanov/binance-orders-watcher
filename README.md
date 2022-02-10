# Binance orders watcher

Web service to watch your binance orders and create alerts.

## Setup

### .env file

`.env` file is used to configure application:
```shell
APP_SCHEMA=                         # http/https
APP_PORT=                           # application port
APP_URI=                            # application public URI
APP_TLS_CERT_PATH=                  # cert.pem path to configure TLS
APP_TLS_KEY_PATH=                   # key.pem path to configure TLS
BINANCE_API_KEY=                    # your Binance account API KEY
BINANCE_API_SECRET=                 # your Binance account API SECRET
BINANCE_PRODUCTION_URI=             # binance prod URI, default should be https://api.binance.com
BASE_AUTH_USERNAME=                 # username for basic authentication
BASE_AUTH_PASSWORD=                 # password for basic authentication
BASE_AUTH_SECRET=                   # secret for basic authentication
MAILJET_API_KEY=                    # your Mailjet account API KEY for alerts
MAILJET_API_SECRET=                 # your Mailjet account API SECRET for alerts
MAILJET_SENDER_NAME=                # your Mailjet account sender name
MAILJET_SENDER_EMAIL=               # your Mailjet account sender email
```

### Docker

To run application in docker perform next steps:

1. generate cert and pem files for TLS configuration and put them to `./tls` directory
2. create `.env` file and fill all fields
3. build the Docker image
    ```shell
      docker build -t binancewatcher .
    ```
4. run the Docker container
    ```shell
      docker run -d --name watcher -p 443:443 binancewatcher
    ```
