package connector

import (
	"context"
	"testing"
)

func TestGetRedisConnection(t *testing.T) {
	GetRedisConnectionWithAuth(context.Background())
}
