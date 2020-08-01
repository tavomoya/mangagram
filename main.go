package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tavomoya/mangagram/actions"
	"github.com/tavomoya/mangagram/models"

	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	tb "gopkg.in/tucnak/telebot.v2"
)

func getMongoClient(conn string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conn))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	db := client.Database("mangagram")
	return db, nil
}

func main() {
	log.Println("Started Manga Gram bot")

	port := os.Getenv("PORT")
	publicURL := os.Getenv("PUBLIC_URL")
	token := os.Getenv("TOKEN")

	conn := os.Getenv("DB_CONN_STRING")

	if conn == "" {
		conn = "mongodb://localhost:27017"
	}

	if port == "" {
		port = "9000"
	}

	listen := fmt.Sprintf(":%s", port)

	db, err := getMongoClient(conn)
	if err != nil {
		log.Fatal("There was an error connecting to DB: ", err)
	}

	dbConfig := &models.DatabaseConfig{
		ConnectionString: conn,
		MongoClient:      db,
	}

	webhook := &tb.Webhook{
		Listen:   listen,
		Endpoint: &tb.WebhookEndpoint{PublicURL: publicURL},
	}

	settings := tb.Settings{
		Token:  token,
		Poller: webhook,
	}

	bot, err := tb.NewBot(settings)
	if err != nil {
		log.Fatal("there was an error creating the bot: ", err)
	}

	jobs := &models.Job{
		Cron: cron.New(),
		DB:   dbConfig,
	}

	// Run Jobs
	actions.GetMangaUpdates(jobs, bot)

	// Available commands:

	bot.Handle("/start", func(m *tb.Message) {

		msg := `
		Hi! This is MangaGram, a Telegram bot for alerts on your favorite manga titles.

		<b>Available Commmands:</b>
		/manga {title} - Get a list of mangas that match the title
		/subscriptions - Get a list of the chat's current manga subscriptions
		/setfeed - Change manga feed used for manga searches (defaults to Manga Reader)
		/help - Info about available commands and mangafeeds
		
		<b>Manga Feeds</b>
		Currently MangaGram supplies manga results from the following pages:

		- Manga Reader (http://manga-reader.fun)
		- Manganelo (https://manganelo.com)
		- Mangaeden (https://mangaeden.com)
		- Kissmanga (https://kissmanga.com)

		You can set your favorite one using the /setfeed command.
		
		If you need help use the /help command.

		MangaGram v0.1.0. Made with ❤️ by @tavomoya.
		`

		_, err := bot.Send(m.Chat, msg, tb.ModeHTML, tb.NoPreview)
		if err != nil {
			log.Println("There was an error sending start msg: ", err)
			return
		}
	})

	bot.Handle("/manga", func(m *tb.Message) {

		name := m.Payload

		if name == "" {
			bot.Send(m.Chat, "<b>No manga name supplied</b>", tb.ModeHTML)
		}

		feedSrc := actions.GetChatMangaFeed(dbConfig, m.Chat.ID)
		if feedSrc == 0 {
			// Do something I guess?
		}

		feed := actions.NewMangaInterface(feedSrc, dbConfig)

		res := feed.QueryManga(name)
		if res == nil || len(res.Suggestions) == 0 {
			bot.Send(m.Chat, "No Manga found with your criteria")
			return
		}

		msg := "These are the manga I found:\n"

		inlineKb := [][]tb.InlineButton{}

		for i, item := range res.Suggestions {
			inlineBtn := []tb.InlineButton{
				{
					Text:   item.Value + " 📖",
					Unique: item.Data,
					URL:    fmt.Sprintf(feed.ViewManga(), item.Data),
				},
				{
					Text:   "Subscribe" + " 🔔",
					Unique: strconv.Itoa(i),
				},
			}

			bot.Handle(&inlineBtn[1], func(btnCb *tb.Callback) {
				fmt.Println("Subscribing user: ", btnCb.Sender.FirstName, inlineBtn[1].Unique, inlineBtn[0].Text, m.Chat.ID)

				// Call the subscribe method of the feed
				manganame := strings.Replace(inlineBtn[0].Text, " 📖", "", 1)
				mangaurl := fmt.Sprintf(feed.ViewManga(), inlineBtn[0].Unique)

				sub := &models.Subscription{
					UserID:    btnCb.Sender.ID,
					UserName:  btnCb.Sender.FirstName,
					ChatID:    m.Chat.ID,
					MangaName: manganame,
					MangaURL:  mangaurl,
				}

				err = feed.Subscribe(sub)
				if err != nil {
					log.Fatal("There was an error subscribing user: ", err)
				}

				bot.Respond(btnCb, &tb.CallbackResponse{
					Text:      "Succesfully subscribed",
					ShowAlert: true,
				})
			})

			inlineKb = append(inlineKb, inlineBtn)
		}

		fmt.Println("Final message and keyboard: ", msg, inlineKb)
		_, err = bot.Send(m.Chat, msg, &tb.ReplyMarkup{
			InlineKeyboard: inlineKb,
		})
		if err != nil {
			log.Fatal("This is not working. Unable to send messages: ", err)
		}
	})

	bot.Handle("/subscriptions", func(m *tb.Message) {

		// Get Chat Subscriptions
		subs, err := actions.GetChatSubscriptions(dbConfig, m.Chat.ID)
		if err != nil {
			log.Println(err)
		}

		if subs == nil || len(subs) == 0 {
			bot.Send(m.Chat, "<b>You're not subscribed to any mangas yet.</b>", tb.ModeHTML)
			return
		}

		btns := [][]tb.InlineButton{}

		for _, s := range subs {

			btn := []tb.InlineButton{
				{
					Text:   s.MangaName + " 📖",
					Unique: s.ID.String(),
					URL:    s.MangaURL,
				},
				{
					Text:   "Remove ❌",
					Unique: s.ID.Hex(),
				},
			}

			bot.Handle(&btn[1], func(btnCb *tb.Callback) {
				err = actions.RemoveMangaSubscription(dbConfig, btn[1].Unique)
				if err != nil {
					log.Fatal("There was an error removing subscription: ", err)
				}

				bot.Respond(btnCb, &tb.CallbackResponse{
					Text:      "Subscription removed",
					ShowAlert: true,
				})
			})

			btns = append(btns, btn)
		}

		_, err = bot.Send(m.Chat, "Current Subscriptions:\n", &tb.ReplyMarkup{
			InlineKeyboard: btns,
		})
		if err != nil {
			log.Fatal("Unable to respond: ", err)
		}
	})

	bot.Handle("/setfeed", func(m *tb.Message) {

		message := "Select feed:\n\n<b>Keep in mind that selecting a different feed than the one you have will remove any current manga subscriptions</b>"

		btns := [][]tb.InlineButton{}
		for _, feed := range actions.AvailableFeeds {

			btn := []tb.InlineButton{
				{
					Text:   feed.Name + "",
					Unique: strconv.Itoa(feed.Code),
				},
				{
					Text:   "🌐",
					Unique: feed.URL,
					URL:    feed.URL,
				},
			}

			bot.Handle(&btn[0], func(btnCb *tb.Callback) {
				c, _ := strconv.Atoi(btn[0].Unique)
				f := models.MangaFeed{
					Code: c,
					URL:  btn[1].Unique,
				}
				err = actions.AddFeedSubscription(dbConfig, m.Chat.ID, f)
				if err != nil {
					log.Fatal("There was an error adding feed subscription: ", err)
				}

				bot.Respond(btnCb, &tb.CallbackResponse{
					Text:      "Feed changed",
					ShowAlert: true,
				})
			})

			btns = append(btns, btn)
		}

		_, err = bot.Send(m.Chat, message, &tb.ReplyMarkup{
			InlineKeyboard: btns,
		}, tb.ModeHTML)
		if err != nil {
			log.Fatal("Unable to respond: ", err)
		}
	})

	bot.Handle("/help", func(m *tb.Message) {
		msg := `
		<b>Available Commmands:</b>
		/manga {title} - Get a list of mangas that match the title
		/subscriptions - Get a list of the chat's current manga subscriptions
		/setfeed - Change manga feed used for manga searches (defaults to Manga Reader)
		/help - Info about available commands and mangafeeds
		
		<b>Manga Feeds</b>
		- Manga Reader (http://manga-reader.fun)
		- Manganelo (https://manganelo.com)
		- Mangaeden (https://mangaeden.com)
		- Kissmanga (https://kissmanga.com)

		You can set your favorite one using the /setfeed command.
		`
		_, err := bot.Send(m.Chat, msg, tb.ModeHTML, tb.NoPreview)
		if err != nil {
			log.Println("There was an error sending start msg: ", err)
			return
		}
	})

	bot.Start()

}
