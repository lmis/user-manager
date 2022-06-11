package db

import (
	"context"
	"time"
)

func DefaultQueryContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
