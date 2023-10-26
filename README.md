# cms-utils-go
This repository contains a collection of utility functions for Go (Golang) development. The utility functions provided in this package are designed to simplify common tasks and enhance the development process. It includes features such as accessing a Redis cluster client with authentication and various other helpful utilities.

## Features
- AWS resource integration
    - lambda
    - S3
    - SES
    - SQS
    - Secret manager
- Elastic Search
- Opensearch
- Database
    - postgres
    - redis
    - redis-cluster with Auth
- slack
- logger
    - go std log
    - logrus
    - slog

## Installation

```shell
go get github.com/happay/cms-utils-go/v2@v2.0.3
go mod tidy
```
Please make sure we update the exiting import 
```go
import github.com/happay/cms-utils-go
```
with major version 2 in order to use redis-cluster client with auth  mode and various other helpful utilities.
```go 
import github.com/happay/cms-utils-go/v2
```
Note: Since the latest version 2 support upgreded version of redis v9, to support this following changes will be required in exiting code:

1. install and update the import with (github.com/redis/go-redis/v9)
```go 
go get github.com/redis/go-redis/v9
```
2. version 9 of redis does not support (*redis.Client).Context() in order to resolve that please use go context (context.Context). eg

Then
```go
func SetRedisKey(key string, data string, exp time.Duration) error {
	return redisClient.Set(redisClient.Context(), key, data, exp).Err()
}
```
Now
```go 
func (r *RedisClient) SetRedisKey(ctx context.Context, key, data string, exp time.Duration) error {
	return r.Set(ctx, key, data, exp).Err()
}
```
### Quickstart
```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/happay/cms-utils-go/v2/connector"
    "github.com/redis/go-redis/v9"
)

func main() {
    // Initialize a Redis client using the utility provided in this package
    ctx := context.Background()
    redisClient := connector.GetRedisConnectionWithAuth(ctx)

    // Set a Redis key with data and expiration time
    key := "myKey"
    data := "myData"
    expiration := 10 * time.Second

    err := redisClient.SetRedisKey(ctx, key, data, expiration)
    if err != nil {
        fmt.Println("Failed to set Redis key:", err)
    } else {
        fmt.Println("Redis key set successfully.")
    }
}

```
## Required enviroment variable
- GRAYLOG_URL: To use logrus logging set GRAYLOG_URL in your env file to get logs in graylog server.
- REDIS_URL: To use the redi-cluster auth connection set REDIS_URL in your env file.
- REDIS_PASSWORD: To use the redi-cluster auth connection set REDIS_PASSWORD in your env file.
