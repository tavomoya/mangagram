package actions

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tavomoya/mangagram/models"

	"go.mongodb.org/mongo-driver/bson"
	tb "gopkg.in/tucnak/telebot.v2"
)

// GetMangaUpdates function runs a goroutine every 6h.
// The goroutine queries the subscription collection and looks
// for new chapters. If a new chapter is found, a message
// is sent to the Chat that got subscribed to the title.
func GetMangaUpdates(job *models.Job, bot *tb.Bot) {
	jobName := "GetMangaUpdates"

	for t := range time.NewTicker(6 * time.Hour).C {
		log.Println("Running Manga Updates Goroutine...", t)
		go func() {
			started := time.Now()
			cursor, err := job.DB.MongoClient.Collection("subscription").Find(job.DB.Ctx, bson.M{})
			if err != nil {
				onError(jobName, started, err)
			}

			subs := make([]*models.Subscription, 0)

			err = cursor.All(job.DB.Ctx, &subs)
			if err != nil {
				onError(jobName, started, err)
			}

			for _, manga := range subs {
				feed := NewMangaInterface(manga.MangaFeed, job.DB)
				if feed == nil {
					continue
				}

				// Get the last chapter for each manga
				last, err := feed.GetLastMangaChapter(manga.MangaURL)
				if err != nil || last == "" || last == fmt.Sprintf(feed.ViewManga(), "") {
					continue // LAter will decide what to do here
				}

				if manga.LastChapterURL == "" || last != manga.LastChapterURL {
					manga.LastChapterURL = last
					msg := fmt.Sprintf("Here is a new chapter for %s\n %s", manga.MangaName, last)
					to, _ := bot.ChatByID(strconv.FormatInt(manga.ChatID, 10))
					bot.Send(to, msg)
					updateLastChapter(manga, job)
				}
			}
		}()
	}
}

func onError(name string, started time.Time, err error) {
	ended := time.Now()
	fmt.Printf("*** [*] Goroutine '%s' finished unexpectedly ***", name)
	fmt.Printf("*** [*] Goroutine '%s' Errors: [%v] ***\n", name, err)
	fmt.Printf("*** [*] Goroutine '%s' end time: %v ***\n", name, ended)
	fmt.Printf("*** [*] Goroutine '%s' time elapsed: %v ***\n", name, ended.Sub(started))
}

func onSuccess(name string, started time.Time) {
	ended := time.Now()
	fmt.Printf("*** [*] Goroutine '%s' finished succesfully ***", name)
	fmt.Printf("*** [*] Goroutine '%s' end time: %v ***\n", name, ended)
	fmt.Printf("*** [*] Goroutine '%s' time elapsed: %v ***\n", name, ended.Sub(started))
}

func updateLastChapter(manga *models.Subscription, job *models.Job) {
	_, err := job.DB.MongoClient.Collection("subscription").UpdateOne(
		job.DB.Ctx,
		bson.M{"_id": manga.ID},
		bson.M{
			"$set": manga,
		},
	)
	if err != nil {
		log.Println("There was an error updating subscription: ", err)
		// return err
	}

	return
}
