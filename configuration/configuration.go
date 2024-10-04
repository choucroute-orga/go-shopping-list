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
	LogLevel            logrus.Level
	DBAddr              string
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
		conf.LogLevel = logrus.InfoLevel
	}

	if logLevel == "debug" {
		conf.LogLevel = logrus.DebugLevel
	} else if logLevel == "error" {
		conf.LogLevel = logrus.ErrorLevel
	} else if logLevel == "info" {
		conf.LogLevel = logrus.InfoLevel
	} else if logLevel == "trace" {
		conf.LogLevel = logrus.TraceLevel
	} else if logLevel == "warn" {
		conf.LogLevel = logrus.WarnLevel
	}

	conf.ListenPort = os.Getenv("API_PORT")
	conf.ListenAddress = os.Getenv("API_ADDRESS")
	conf.ListenRoute = os.Getenv("API_ROUTE")

	conf.DBAddr = os.Getenv("REDIS_ADDR")
	conf.DBPassword = os.Getenv("REDIS_PASSWORD")

	conf.RabbitURI = os.Getenv("RABBITMQ_URL")

	conf.TranslateValidation, err = strconv.ParseBool(os.Getenv("TRANSLATE_VALIDATION"))

	if err != nil {
		logger.Error("Failed to parse bool for TRANSLATE_VALIDATION")
		os.Exit(1)
	}

	conf.JWTSecret = os.Getenv("JWT_SECRET")

	return &conf
}
