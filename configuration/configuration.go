package configuration

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"context": "configuration/configuration",
})

type Configuration struct {
	ListenPort          string
	ListenAddress       string
	ListenRoute         string
	LogLevel            string
	DBHost              string
	DBPort              string
	DBName              int
	DBPassword          string
	TranslateValidation bool
	RabbitURI           string
	JWTSecret           string
}

func New() *Configuration {

	conf := Configuration{}
	var err error

	logLevel := os.Getenv("LOG_LEVEL")
	if len(logLevel) < 1 || logLevel != "debug" && logLevel != "error" && logLevel != "info" && logLevel != "trace" && logLevel != "warn" {
		logrus.WithFields(logrus.Fields{
			"logLevel": logLevel,
		}).Info("logLevel not conform, use `info` ")
		conf.LogLevel = "info"
	} else {
		conf.LogLevel = logLevel
	}

	conf.ListenPort = os.Getenv("API_PORT")
	conf.ListenAddress = os.Getenv("API_ADDRESS")
	conf.ListenRoute = os.Getenv("API_ROUTE")

	conf.DBHost = os.Getenv("REDIS_HOST")
	conf.DBPort = os.Getenv("REDIS_PORT")
	// Convert to int
	conf.DBName, err = strconv.Atoi(os.Getenv("REDIS_DATABASE"))
	if err != nil {
		logger.Error("Failed to parse int for REDIS_DATABASE")
		os.Exit(1)
	}
	conf.DBPassword = os.Getenv("REDIS_PASSWORD")

	rabbitPort := os.Getenv("RABBITMQ_PORT")
	rabbitUser := os.Getenv("RABBITMQ_DEFAULT_USER")
	rabbitPassword := os.Getenv("RABBITMQ_DEFAULT_PASS")
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	conf.RabbitURI = "amqp://" + rabbitUser + ":" + rabbitPassword + "@" + rabbitHost + ":" + rabbitPort + "/"

	conf.TranslateValidation, err = strconv.ParseBool(os.Getenv("TRANSLATE_VALIDATION"))

	if err != nil {
		logger.Error("Failed to parse bool for TRANSLATE_VALIDATION")
		os.Exit(1)
	}

	conf.JWTSecret = os.Getenv("JWT_SECRET")

	return &conf
}
