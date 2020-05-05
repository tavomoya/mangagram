package models

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// DatabaseConfig struct defines
// some properties that are used
// for databse coonfiguration.
type DatabaseConfig struct {
	ConnectionString string
	MongoClient      *mongo.Database
	Ctx              context.Context
}
