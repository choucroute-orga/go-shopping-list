package tests

import (
	"context"
	"fmt"
	"log"
	"shopping-list/configuration"

	dockertest "github.com/ory/dockertest/v3"
	"github.com/redis/go-redis/v9"
)

// InitTestDocker function initialize docker with mongo image used for integration tests
func InitTestDocker(exposedPort string) (*redis.Client, *dockertest.Pool, *dockertest.Resource) {
	var rdb *redis.Client
	var err error
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.Run("bitnami/redis", "latest", []string{"ALLOW_EMPTY_PASSWORD=yes"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		rdb = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp")),
		})

		return rdb.Ping(context.Background()).Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return rdb, pool, resource
}

func CloseTestDocker(client *redis.Client, pool *dockertest.Pool, resource *dockertest.Resource) {
	// When you're done, kill and remove the container
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	client.Conn().Close()
}

func GetDefaultConf() *configuration.Configuration {
	return &configuration.Configuration{
		ListenAddress: "localhost",
		ListenPort:    "3000",
		DBHost:        "localhost",
		DBPort:        "6379",
	}
}
