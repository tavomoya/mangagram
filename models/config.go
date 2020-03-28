package models

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type DatabaseConfig struct {
	ConnectionString string
	MongoClient      *mongo.Database
	Ctx              context.Context
}
