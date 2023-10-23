package connector

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"sync"

	"github.com/happay/cms-utils-go/logger"
	"github.com/redis/go-redis/v9"

	"github.com/happay/cms-utils-go/util"
)

const Pong = "PONG"

var redisClient *redis.Client
var redisClusterClient *redis.ClusterClient
var redisConn sync.Once

// GetRedisConn returns the redis.Client object. It takes a parameter redisAddrKey
// which is the Redis env variable key stored on the os or AWS Parameter Store.
// This function will try to get the value of the input key from the os or Parameter Store and create the redisClient.
func GetRedisConn(redisAddrKey string) *redis.Client {
	redisConn.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     util.GetConfigValue(redisAddrKey),
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		ctx := context.Background()
		pong, err := redisClient.Ping(ctx).Result()
		if err != nil || pong != Pong {
			reason := fmt.Sprintf("Error while creating Redis connection pool: %s", err)
			logger.GetLogger("", "").Println(reason)
		}
	})
	return redisClient
}

var createRedisConnection sync.Once

const (
	RedisGet    = "redis/get"
	RedisSet    = "redis/set"
	RedisExist  = "redis/exists"
	RedisDelete = "redis/delete"
)

func GetRedisConnectionWithAuth(ctx context.Context) *redis.ClusterClient {
	createRedisConnection.Do(func() {
		redisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    []string{os.Getenv("REDIS_URL")},
			Password: os.Getenv("REDIS_PASSWORD"),
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},

			ReadOnly:       false,
			RouteRandomly:  false,
			RouteByLatency: false,
		})
		pong, err := redisClusterClient.Ping(ctx).Result()
		if err != nil || pong != "PONG" {
			reason := fmt.Sprintf("Error while creating Redis connection pool: %s", err)
			logger.GetLogger("", "").Println(reason)
		}
	})
	return redisClusterClient
}
