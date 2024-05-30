package main

import (
	"fmt"
	"shopping-list/api"
	"shopping-list/configuration"
	"shopping-list/db"
	"shopping-list/validation"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"context": "main",
})

func main() {
	logger.Info("Shopping List API Starting...")

	conf := configuration.New()

	rdb := db.New(conf)

	val := validation.New(conf)
	r := api.New(val)
	v1 := r.Group(conf.ListenRoute)

	h := api.NewApiHandler(conf, rdb)

	h.Register(v1, conf)
	r.Logger.Fatal(r.Start(fmt.Sprintf("%v:%v", conf.ListenAddress, conf.ListenPort)))

	defer rdb.Close()
}
