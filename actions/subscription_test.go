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

func TestRemoveMangaSubscription(t *testing.T) {
	opts := &mtest.Options{}
	opts.ClientType(mtest.Mock)
	opts.CollectionName("subscription")
	opts.DatabaseName("mangagram")
	opts.ShareClient(true)

	mt := mtest.New(t, opts)
	defer mt.Close()

	is := is.New(t)
	config := &models.DatabaseConfig{
		Ctx:         context.Background(),
		MongoClient: mt.Client.Database("mangagram"),
	}

	mt.Run("Nil Database", func(t *mtest.T) {
		err := RemoveMangaSubscription(nil, "1")
		is.True(err != nil)
	})

	mt.Run("Invalid ObjectID", func(t *mtest.T) {
		err := RemoveMangaSubscription(config, "abc123")
		is.True(err != nil)
	})

	mt.Run("Failed to delete", func(t *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 0}})
		err := RemoveMangaSubscription(config, "60fc82d3188b85f46f5f6b9c")
		is.True(err != nil)
	})

	mt.Run("Success", func(t *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 1}})
		err := RemoveMangaSubscription(config, "60fc82d3188b85f46f5f6b9c")
		is.NoErr(err)
	})
}

func TestGetChatMangaFeed(t *testing.T) {
	opts := &mtest.Options{}
	opts.ClientType(mtest.Mock)
	opts.CollectionName("feed_sub")
	opts.DatabaseName("mangagram")
	opts.ShareClient(true)

	mt := mtest.New(t, opts)
	defer mt.Close()

	is := is.New(t)
	config := &models.DatabaseConfig{
		Ctx:         context.Background(),
		MongoClient: mt.Client.Database("mangagram"),
	}

	mt.Run("Nil Database", func(t *mtest.T) {
		feed := GetChatMangaFeed(nil, 1)
		is.Equal(0, feed)
	})

	mt.Run("Failed to query", func(t *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code: 90,
		}))

		feed := GetChatMangaFeed(config, 1)
		is.Equal(0, feed)
	})

	mt.Run("Success", func(t *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "feed_sub.code", mtest.FirstBatch, bson.D{
			{"_id", primitive.NewObjectID()},
			{"code", 100},
		}))

		feed := GetChatMangaFeed(config, 1)
		is.Equal(100, feed)
	})
}

func TestAddFeedSubscription(t *testing.T) {
	opts := &mtest.Options{}
	opts.ClientType(mtest.Mock)
	opts.CollectionName("feed_sub")
	opts.DatabaseName("mangagram")
	opts.ShareClient(true)

	mt := mtest.New(t, opts)
	defer mt.Close()

	is := is.New(t)
	config := &models.DatabaseConfig{
		Ctx:         context.Background(),
		MongoClient: mt.Client.Database("mangagram"),
	}

	mt.Run("Nil Database", func(t *mtest.T) {
		err := AddFeedSubscription(nil, 0, models.MangaFeed{})
		is.True(err != nil)
	})

	mt.Run("Invalid chat ID", func(t *mtest.T) {
		err := AddFeedSubscription(config, 0, models.MangaFeed{})
		is.True(err != nil)
	})

	mt.Run("Failed to insert new feed", func(t *mtest.T) {
		feed := models.MangaFeed{
			Code: 1,
			Name: "manga test",
			URL:  "http://mangatest.test",
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "feed_sub.chatid", mtest.FirstBatch))

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Code:    11000,
			Message: "something went wrong",
		}))

		err := AddFeedSubscription(config, 10, feed)
		is.True(err != nil)
	})

	mt.Run("Success inserting new feed", func(t *mtest.T) {
		feed := models.MangaFeed{
			Code: 1,
			Name: "manga test",
			URL:  "http://mangatest.test",
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "feed_sub.chatid", mtest.FirstBatch))

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := AddFeedSubscription(config, 10, feed)
		is.NoErr(err)
	})

	mt.Run("Failed to update feed", func(t *mtest.T) {
		feed := models.MangaFeed{
			Code: 1,
			Name: "manga test",
			URL:  "http://mangatest.test",
		}
		id := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "feed_sub.chatid", mtest.FirstBatch, bson.D{
			{"_id", id},
			{"url", "oldurl"},
			{"code", 10},
			{"chatId", 10},
		}), mtest.CreateCursorResponse(0, "feed_sub.chatid", mtest.NextBatch))

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Code:    11000,
			Message: "something went wrong",
		}))

		err := AddFeedSubscription(config, 10, feed)
		is.True(err != nil)
	})

	mt.Run("Success update feed", func(t *mtest.T) {
		feed := models.MangaFeed{
			Code: 1,
			Name: "manga test",
			URL:  "http://mangatest.test",
		}
		id := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "feed_sub.chatid", mtest.FirstBatch, bson.D{
			{"_id", id},
			{"url", "oldurl"},
			{"code", 10},
			{"chatId", 10},
		}), mtest.CreateCursorResponse(0, "feed_sub.chatid", mtest.NextBatch))

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{
				{"_id", id},
				{"url", "http://mangatest.test"},
				{"code", 1},
				{"chatId", 10},
			}},
		})

		err := AddFeedSubscription(config, 10, feed)
		is.NoErr(err)
	})
}
