package connector

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/happay/cms-utils-go/v2/logger"
	"github.com/redis/go-redis/v9"

	"github.com/happay/cms-utils-go/v2/util"
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

func SetRedisKey(key string, data string, exp time.Duration) error {
	ctx := context.Background()
	GetRedisConnectionWithAuth(ctx)
	return redisClusterClient.Set(ctx, key, data, exp).Err()
}

func SetRedisKeyObject(key string, data interface{}, exp time.Duration) error {
	ctx := context.Background()
	GetRedisConnectionWithAuth(ctx)
	return redisClusterClient.Set(ctx, key, data, exp).Err()
}

func GetRedisKeyValueString(key string) (result string, err error) {
	ctx := context.Background()
	GetRedisConnectionWithAuth(ctx)
	if !RedisKeyExists(key) {
		err = fmt.Errorf("key %s doesn't exist", key)
		return
	}
	result = redisClusterClient.Get(ctx, key).Val()
	return
}

func GetRedisKeyValueBytes(key string) (result []byte, err error) {
	ctx := context.Background()
	GetRedisConnectionWithAuth(ctx)
	if !RedisKeyExists(key) {
		err = fmt.Errorf("key %s doesn't exist", key)
		return
	}
	result, err = redisClusterClient.Get(ctx, key).Bytes()
	return
}

func RedisKeyExists(refNo string) bool {
	ctx := context.Background()
	GetRedisConnectionWithAuth(ctx)
	return redisClusterClient.Exists(ctx, refNo).Val() == 1
}

func DeleteRedisKey(refNo string) error {
	ctx := context.Background()
	GetRedisConnectionWithAuth(ctx)
	return redisClusterClient.Del(ctx, refNo).Err()
}

func GetRedisTTL(key string) time.Duration {
	ctx := context.Background()
	GetRedisConnectionWithAuth(ctx)
	return redisClusterClient.TTL(ctx, key).Val()
}
