package config

import "github.com/spf13/viper"

type Config struct {
	AppPort            string `mapstructure:"APP_PORT"`
	AppURI             string `mapstructure:"APP_URI"`
	AppSchema          string `mapstructure:"APP_SCHEMA"`
	AppTlsCertPath     string `mapstructure:"APP_TLS_CERT_PATH"`
	AppTlsKeyPath      string `mapstructure:"APP_TLS_KEY_PATH"`
	BinApiKey          string `mapstructure:"BINANCE_API_KEY"`
	BinApiSecret       string `mapstructure:"BINANCE_API_SECRET"`
	BinProdURI         string `mapstructure:"BINANCE_PRODUCTION_URI"`
	BaseAuthUsername   string `mapstructure:"BASE_AUTH_USERNAME"`
	BaseAuthPassword   string `mapstructure:"BASE_AUTH_PASSWORD"`
	BaseAuthSecret     string `mapstructure:"BASE_AUTH_SECRET"`
	MailjetApiKey      string `mapstructure:"MAILJET_API_KEY"`
	MailjetApiSecret   string `mapstructure:"MAILJET_API_SECRET"`
	MailjetSenderName  string `mapstructure:"MAILJET_SENDER_NAME"`
	MailjetSenderEmail string `mapstructure:"MAILJET_SENDER_EMAIL"`
}

func New(path string, name string) (config *Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
