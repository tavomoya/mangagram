package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Subscription is a struct used to define
// a manga subscription.
type Subscription struct {
	// Internal ID assiged by MongoDB
	ID primitive.ObjectID `bson:"_id"`

	// Id of the user who added the subscription
	UserID int

	// User name of the user who added the subscription
	UserName string

	// Name of the user who added the subscription
	FirstName string

	// ID of the chat the susbcription is being added on
	ChatID int64

	// Name of the manga
	MangaName string

	// URL to the manga
	MangaURL string

	// URL to the last chapter published for the manga
	LastChapterURL string

	// Feed this subscription belongs to
	MangaFeed int
}

// FeedSubs is a struct used to define
// a subscription to a manga feed.
type FeedSubs struct {
	// Internal ID assigned by MongoDB
	ID primitive.ObjectID `bson:"_id"`

	// URL of the Manga Feed
	URL string

	// Internal Code of the Manga Feed
	Code int

	// ID of the Chat subscribed
	ChatID int64

	// Id of the user subscribed (unused)
	UserID int
}
