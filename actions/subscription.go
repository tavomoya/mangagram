package actions

import (
	"errors"
	"log"
	"mangagram/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetChatSubscriptions(db *models.DatabaseConfig, chatID int64) ([]*models.Subscription, error) {

	if db == nil {
		log.Println("The DB model is nil")
		return nil, errors.New("The DB model passed is nil, can't operate")
	}

	cursor, err := db.MongoClient.Collection("subscription").Find(db.Ctx, bson.M{"chatid": chatID})
	if err != nil {
		log.Println("There was an error trying to look for this chat's subscriptions: ", err)
		return nil, err
	}

	subs := make([]*models.Subscription, 0)
	err = cursor.All(db.Ctx, &subs)
	if err != nil {
		log.Println("There was an error trying to decode subscriptions into a subscriptions slice: ", err)
		return nil, err
	}

	return subs, nil
}

func RemoveMangaSubscription(db *models.DatabaseConfig, subscriptionID string) error {

	if db == nil {
		log.Println("The DB model is nil")
		return errors.New("The DB model passed is nil, can't operate")
	}

	id, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		log.Println("Invalid subscription ID: ", subscriptionID)
		return err
	}

	res, err := db.MongoClient.Collection("subscription").DeleteOne(db.Ctx, bson.M{"_id": id})
	if err != nil {
		log.Println("There was an error trying to remove subscription: ", err)
		return err
	}

	if res.DeletedCount != 1 {
		log.Println("Couldn't delete subscription: ", res.DeletedCount)
		return errors.New("An unexpected error happened and the subscription was not deleted.")
	}

	return nil
}

func GetChatMangaFeed(db *models.DatabaseConfig, chatID int64) int {

	if db == nil {
		log.Println("The DB model is nil")
		return 0
	}

	feed := models.FeedSubs{}
	res := db.MongoClient.Collection("feed_sub").FindOne(db.Ctx, bson.M{"chatid": chatID})
	err := res.Decode(&feed)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("Did not find any feed subs for this chat. Returning default feed")
			return 1
		}
		log.Println("There was an unexpected error decoding feed_sub document into a struct: ", err)
		return 0
	}

	return feed.Code
}

func AddFeedSubscription(db *models.DatabaseConfig, chatID int64, feed models.MangaFeed) error {

	if db == nil {
		log.Println("The DB model is nil")
		return errors.New("The DB model passed is nil, can't operate")
	}

	// Validate chat ID
	if chatID == 0 {
		log.Println("ChatID can't be 0")
		return errors.New("No Chat supplied for feed subscription")
	}

	sub := models.FeedSubs{
		ID:     primitive.NewObjectID(),
		URL:    feed.URL,
		Code:   feed.Code,
		ChatID: chatID,
	}

	// Check if it exists:
	f := models.FeedSubs{}
	res := db.MongoClient.Collection("feed_sub").FindOne(db.Ctx, bson.M{"chatid": chatID})
	err := res.Decode(&f)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			_, err = db.MongoClient.Collection("feed_sub").InsertOne(db.Ctx, sub)
			if err != nil {
				log.Println("There was an error saving the feed sub: ", err)
				return err
			}

			return nil
		}
		log.Println("There was an unexpected error decoding feed_sub document into a struct: ", err)
		return err
	}

	f.Code = feed.Code
	f.URL = feed.URL
	_, err = db.MongoClient.Collection("feed_sub").UpdateOne(db.Ctx, bson.M{"_id": f.ID}, bson.M{"$set": f})
	if err != nil {
		log.Println("There was an error saving the feed sub: ", err)
		return err
	}

	return nil
}
