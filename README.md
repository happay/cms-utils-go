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
    - mysql
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
go get github.com/happay/cms-utils-go/v3@v2.0.3
go mod tidy
```
Please make sure we update the exiting import 
```go
import github.com/happay/cms-utils-go
```
with major version 2 in order to use redis-cluster client with auth  mode and various other helpful utilities.
```go 
import github.com/happay/cms-utils-go/v3
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

    "github.com/happay/cms-utils-go/v3/connector"
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

	"github.com/happay/cms-utils-go/v3/tracing"
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
	utilsMiddleware "github.com/happay/cms-utils-go/v3/middleware"
)
func main(){
    r := gin.Default()
	r.Use(utilsMiddleware.TracingMiddleware("my-service-name"))
    r.Run()
}
    
```

To create a span with global tracer provider

```
import (
    "github.com/happay/cms-utils-go/v3/tracing"
)
func HttpCall(ctx context.Context) {
	_, span := tracing.StartSpanWithGlobalTracer(ctx, "HttpCall")
	// Do something
}
```

To create a span with expicilty pass the tracer provider
```
import (
    "github.com/happay/cms-utils-go/v3/tracing"
    ...
)

var DDProvider tracing.DataDogProvider


func init() {
	ddConfig := &tracing.DataDogTracerConfig{
		ServiceName: AppName,
		Env:         os.Getenv("APP_ENV"),
		Version:     "v1",
		Host:        os.Getenv("DD_AGENT_HOST"),
		Port:        os.Getenv("DD_DOGSTATSD_PORT"),
	}
	DDProvider = tracing.DataDogProvider{
		TracerConfig: ddConfig,
	}
}

func main() {
    DDProvider.NewTracerProvider()
	defer func() {
		if err := DDProvider.TracerProvider.Shutdown(); err != nil {
			common.Errorf("Error shutting down tracer provider: %v", err)
		}
		common.Infof("tracer provider shutting down gracefully")
	}()
	DDProvider.InitTracing()
	// Start your service
}

func HttpCall(ctx context.Context) {
	_, span := DDProvider.StartSpan(ctx, "HttpCall")
	// Do something
}
```

In above code, on init function we set the ddprovider value and in main function the value get set as tracer provider, which we can use as global variable.

# Database
## Mysql Connection

To connect with mysql connection 
```
type Mysql struct {
	Host               string `json:"host"`
	Port               string `json:"port"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	Database           string `json:"database"`
	MaxOpenConnections string `json:"maxOpenConnections"`
	MaxIdleConnections string `json:"maxIdleConnections"`
}
```
use above Mysql map to pass `mysqlConfig`

```
db = connector.GetMySqlConn(mysqlConfig, "mysql")
```

# Utils
## Http call

Function `MakeHttpRequest` provides the http call to network, which take http path, [http method](https://pkg.go.dev/net/http#pkg-constants) and `HttpOption` optional parmaters.

### WithQueryParam
`func WithQueryParam(queryParams map[string]string) HttpOption`

This will take all the Query params in map[string]string

### WithRequestBody
`func WithRequestBody(requestBody PropertyMap) HttpOption`

This will set request body to http request

### WithHeader

`func WithHeader(header map[string]string) HttpOption`

This will set header to http request, if http request accept the headers

### WithTimeoutInSec

`func WithTimeoutInSec(timeout int64) HttpOption`

Sets the timeout if required.

### WithCertificate

`func WithCertificate(publicKey, privateKey string) HttpOption`

*NOTE*: PropertyMap is `type PropertyMap map[string]interface{}`

e.g.
```
statusCode, responseBody, err := util.MakeHttpRequest(
		http.MethodPost,
		path,
		util.WithHeader(reqData.Headers),
		util.WithRequestBody(util.PropertyMap(reqData.RequestBody)),
	)
```
