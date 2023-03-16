package connector

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/happay/cms-utils-go/util"
)

const Pong = "PONG"

var redisClient *redis.Client
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
			//logger.GetLogger().Println(reason)
			fmt.Println(reason)
		}
	})
	return redisClient
}
