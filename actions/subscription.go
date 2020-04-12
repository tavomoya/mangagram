package actions

import (
	"errors"
	"log"
	"mangagram/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
