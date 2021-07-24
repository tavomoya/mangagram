package actions

import (
	"context"
	"testing"

	"github.com/matryer/is"
	"github.com/tavomoya/mangagram/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestGetChatSubscriptions(t *testing.T) {
	opts := &mtest.Options{}
	opts.ClientType(mtest.Mock)
	opts.CollectionName("subscription")
	opts.DatabaseName("mangagram")
	opts.ShareClient(true)

	mt := mtest.New(t, opts)
	defer mt.Close()

	is := is.New(t)

	mt.Run("Nil Database", func(t *mtest.T) {
		_, err := GetChatSubscriptions(nil, 1)
		is.True(err != nil)
	})

	mt.Run("Failed query", func(t *mtest.T) {
		config := &models.DatabaseConfig{
			Ctx:         context.Background(),
			MongoClient: mt.Client.Database("mangagram"),
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    90,
			Message: "Basic Error",
		}))

		_, err := GetChatSubscriptions(config, 1)
		is.True(err != nil)
		is.Equal(err.Error(), "Basic Error")
	})

	mt.Run("success query", func(t *mtest.T) {
		config := &models.DatabaseConfig{
			Ctx:         context.Background(),
			MongoClient: mt.Client.Database("mangagram"),
		}

		firstId := primitive.NewObjectID()
		first := mtest.CreateCursorResponse(1, "subscription.chatid", mtest.FirstBatch, bson.D{
			{"_id", firstId},
			{"userName", "jcase"},
			{"chatId", 1},
			{"mangaUrl", "http://mangafeed.com/naruto"},
		})

		secondId := primitive.NewObjectID()
		second := mtest.CreateCursorResponse(1, "subscription.chatid", mtest.NextBatch, bson.D{
			{"_id", secondId},
			{"userName", "jcase"},
			{"chatId", 1},
			{"mangaUrl", "http://mangafeed.com/one-piece"},
		})
		end := mtest.CreateCursorResponse(0, "subscription.chatid", mtest.NextBatch)
		mt.AddMockResponses(first, second, end)

		subs, err := GetChatSubscriptions(config, 1)
		is.NoErr(err)
		is.Equal(subs, []*models.Subscription{
			{
				ID:       firstId,
				UserName: "jcase",
				ChatID:   1,
				MangaURL: "http://mangafeed.com/naruto",
			},
			{
				ID:       secondId,
				UserName: "jcase",
				ChatID:   1,
				MangaURL: "http://mangafeed.com/one-piece",
			},
		})

	})

}
