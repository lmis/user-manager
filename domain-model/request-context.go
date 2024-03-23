package domain_model

import (
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
)

type RequestContext struct {
	RequestID string
	User      User
	Database  *mongo.Database
	Logger    *slog.Logger
	Config    *Config
}
