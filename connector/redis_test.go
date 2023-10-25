package connector

import (
	"context"
	"testing"
)

func TestGetRedisConnection(t *testing.T) {
	GetRedisConnectionWithAuth(context.Background())
}

func TestGetRedisConn(t *testing.T) {
	GetRedisConn(":6379")
}
