package actions

import (
	"fmt"
	"log"
	"mangagram/models"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson"
	tb "gopkg.in/tucnak/telebot.v2"
)

const schedule string = "@every 12h"

func GetMangaUpdates(job *models.Job, bot *tb.Bot) {
	log.Println("Am I running? ", schedule, job, bot)
	job.Cron.AddFunc(schedule, func() {
		jobName := "GetMangaUpdates"
		log.Println("Running Manga Updates CRON")
		started := time.Now()
		fmt.Println("*** [*] CRON job 'GetMangaUpdates' started ***")
		fmt.Printf("*** [*] CRON job 'GetMangaUpdates' start time: %v ***\n", started)

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
			// Get the last chapter for each manga
			last, err := getMangaLastChapter(manga.MangaURL)
			if err != nil {
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

		onSuccess(jobName, started)
	})

	job.Cron.Start()
}

func onError(name string, started time.Time, err error) {
	ended := time.Now()
	fmt.Printf("*** [*] CRON job '%s' finished unexpectedly ***", name)
	fmt.Printf("*** [*] CRON job '%s' Errors: [%v] ***\n", name, err)
	fmt.Printf("*** [*] CRON job '%s' end time: %v ***\n", name, ended)
	fmt.Printf("*** [*] CRON job '%s' time elapsed: %v ***\n", name, ended.Sub(started))
}

func onSuccess(name string, started time.Time) {
	ended := time.Now()
	fmt.Println("*** [*] CRON job '%s' finished succesfully ***", name)
	fmt.Printf("*** [*] CRON job '%s' end time: %v ***\n", name, ended)
	fmt.Printf("*** [*] CRON job '%s' time elapsed: %v ***\n", name, ended.Sub(started))
}

func getMangaLastChapter(titleURL string) (string, error) {

	if titleURL == "" {
		log.Println("No title supplied")
		return "", nil
	}

	page, err := goquery.NewDocument(titleURL)
	if err != nil {
		log.Println("There was an error getting the page: ", err)
		return "", err
	}

	lastChaperUrl, _ := page.Find("a.chapter-name").First().Attr("href")

	return lastChaperUrl, nil
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
