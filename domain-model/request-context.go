package domain_model

import "go.mongodb.org/mongo-driver/mongo"

type RequestContext struct {
	User        User
	Database    *mongo.Database
	Log         Log
	SecurityLog SecurityLog
	Config      *Config
}
