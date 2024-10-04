package main

import (
	"context"
	"fmt"
	"log"
	"shopping-list/api"
	"shopping-list/configuration"
	"shopping-list/db"
	"shopping-list/messages"
	"shopping-list/validation"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"context": "main",
})

func main() {
	configuration.SetupLogging()
	logger.Info("Shopping List API Starting...")

	conf := configuration.New()
	logger.Logger.SetLevel(conf.LogLevel)

	rdb := db.New(conf)

	val := validation.New(conf)
	r := api.New(val)
	v1 := r.Group(conf.ListenRoute)
	amqp := messages.New(conf)
	h := api.NewApiHandler(conf, rdb, amqp)

	h.Register(v1, conf)
	tp, _ := api.InitOtel()

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
		if err := rdb.Close(); err != nil {
			logger.WithError(err).Error("Error closing redis connection")
		}
	}()

	go func() {
		h.ConsumeMessages()
	}()

	r.Logger.Fatal(r.Start(fmt.Sprintf("%v:%v", conf.ListenAddress, conf.ListenPort)))

}
