package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	ID             primitive.ObjectID `bson:"_id"`
	UserID         int
	UserName       string
	FirstName      string
	ChatID         int64
	MangaName      string
	MangaURL       string
	LastChapterURL string
	MangaFeed      int
}

type FeedSubs struct {
	ID     primitive.ObjectID `bson:"_id"`
	URL    string
	Code   int
	ChatID int64
	UserID int
}
