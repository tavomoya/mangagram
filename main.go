package main

import (
	"context"
	"fmt"
	"log"
	"mangagram/actions"
	"mangagram/models"
	"os"
	"strings"
	"time"

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

	bot.Handle("/manga", func(m *tb.Message) {

		name := m.Payload

		if name == "" {
			bot.Send(m.Chat, "<b>No manga name supplied</b>", tb.ModeHTML)
		}

		feed := actions.NewMangaInterface(2, dbConfig)

		res := feed.QueryManga(name)
		if res == nil {
			bot.Send(m.Sender, "No Manga found with your criteria")
		}

		msg := "These are the manga I found:\n"

		inlineKb := [][]tb.InlineButton{}

		for _, item := range res.Suggestions {
			inlineBtn := []tb.InlineButton{
				{
					Text:   item.Value + " 📖",
					Unique: item.Data,
					URL:    fmt.Sprintf(feed.ViewManga(), item.Data),
				},
				{
					Text:   "Subscribe" + " 🔔",
					Unique: item.Data + "_sub",
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

		if subs == nil {
			bot.Send(m.Chat, "<b>You're not subscribed to any mangas yet.</b>", tb.ModeHTML)
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

			fmt.Println("Btn slice: ", btn)
			bot.Handle(&btn[1], func(btnCb *tb.Callback) {
				fmt.Println("Remove sub: ", btn[0].Text, btn[0].Unique, btn[1].Unique)
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

	bot.Start()

	// Testing server

	// router := mux.NewRouter()

	// router.HandleFunc("/manga/{name}", func(w http.ResponseWriter, r *http.Request) {
	// 	mangaName, _ := url.QueryUnescape(mux.Vars(r)["name"])
	// 	log.Println("The name: ", mangaName)
	// 	res := actions.QueryManga(mangaName)
	// 	if res == nil {
	// 		w.WriteHeader(http.StatusNotFound)
	// 		return
	// 	}

	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusOK)
	// 	json.NewEncoder(w).Encode(&res)
	// }).Methods("GET")

	// http.ListenAndServe(listen, handlers.CombinedLoggingHandler(os.Stdout, router))
}
