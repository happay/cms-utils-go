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
- tracing
  - datadog
  - opentelemetry
- middleware
  - tracing

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
## Redis
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

# Enable Tracing
To enable tracing we need to follow the below step or given piece of code.

Configure the datadog config, where will define the below varibales,
  - service Name
  - host
  - port
  - version (optional)
The default is localhost:8126. It should contain both host and port.

```
package main

import (
	"os"
	"../api"
	"../common"
	"../constant"

	"github.com/happay/cms-utils-go/v2/tracing"
)

func main() {
	ddConfig := &tracing.DataDogTracerConfig{
		ServiceName: constant.APP_NAME,
		Env:         os.Getenv("APP_ENV"),
		Host:        os.Getenv("DD_AGENT_HOST"),
		Port:        os.Getenv("DD_DOGSTATSD_PORT"),
	}
	ddProvider := tracing.DataDogProvider{
		TracerConfig: ddConfig,
	}
	ddProvider.NewTracerProvider()
	defer func() {
		if err := ddProvider.TracerProvider.Shutdown(); err != nil {
			common.GetLogger().Error("Error shutting down tracer provider: ", err)
		}
	}()
	ddProvider.InitTracing()
	api.StartServices()
}

```

In above pieace of code, first we define the DataDogTracerConfig, to initialise the datadog agent. InitTracing will set the tracer provider as global tracer and set the propogator which will help us to allow the distributed tracing (cross-cutting concern). 

Once we set the tracer, we can start with incoming requirest tracing. Adding `TracingMiddleware` as middleware to instrumented service we can trace the incoming requests.

```
package main

import (
	"github.com/gin-gonic/gin"
	utilsMiddleware "github.com/happay/cms-utils-go/v2/middleware"
)
func main(){
    r := gin.Default()
	r.Use(utilsMiddleware.TracingMiddleware("my-service-name"))
    r.Run()
}
    
```