package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"shopping-list/api"
	"shopping-list/configuration"
	"shopping-list/db"
	"shopping-list/messages"
	"shopping-list/validation"
	"syscall"

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
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.WithError(err).Error("Error shutting down tracer provider")
		}
		if err := rdb.Close(); err != nil {
			logger.WithError(err).Error("Error closing redis connection")
		}
	}()

	go func() {
		h.ConsumeMessages()
	}()

	go func() {
		h.ConsumeAddIngredientMessage(ctx)
	}()

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down gracefully...")
		cancel()
	}()

	r.Logger.Fatal(r.Start(fmt.Sprintf("%v:%v", conf.ListenAddress, conf.ListenPort)))

}
