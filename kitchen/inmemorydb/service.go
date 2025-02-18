package inmemorydb

import (
	"context"

	"github.com/superryanguo/kitchen/queue"
)

type OrderRequestInMemoryService interface {
	SetOrder(ctx context.Context, key string, request queue.OrderQueueRequest) error
	GetOrder(ctx context.Context) (*queue.OrderQueueRequest, error)
}
