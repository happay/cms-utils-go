package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func TracingMiddleware(serviceName string) func(*gin.Context) {
	return otelgin.Middleware(serviceName)
}
