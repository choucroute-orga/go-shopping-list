package db

import (
	"context"
	"shopping-list/configuration"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func New(configuration *configuration.Configuration) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     configuration.DBHost + ":" + configuration.DBPort,
		Password: configuration.DBPassword,
		DB:       configuration.DBName,
	})

	// Test the connectivity of the Redis server
	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}
	logrus.Debug("Connected to Redis")
	return rdb

}


