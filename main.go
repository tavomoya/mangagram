package main

import (
	"fmt"
	"log"
	"mangagram/actions"
	"os"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	log.Println("Started Manga Gram bot")

	port := os.Getenv("PORT")
	publicUrl := os.Getenv("PUBLIC_URL")
	token := os.Getenv("TOKEN")

	if port == "" {
		port = "9000"
	}

	listen := fmt.Sprintf(":%s", port)

	webhook := &tb.Webhook{
		Listen:   listen,
		Endpoint: &tb.WebhookEndpoint{PublicURL: publicUrl},
	}

	settings := tb.Settings{
		Token:  token,
		Poller: webhook,
	}

	bot, err := tb.NewBot(settings)
	if err != nil {
		log.Fatal("there was an error creating the bot: ", err)
	}

	// Available commands:

	bot.Handle("/manga", func(m *tb.Message) {
		fmt.Println("The message received is ", m.Text)

		name := strings.Replace(m.Text, "/manga ", "", 1)

		if name == "" {
			bot.Send(m.Sender, "No manga name supplied")
		}

		feed := actions.NewMangaInterface(1)

		res := feed.QueryManga(name)
		if res == nil {
			bot.Send(m.Sender, "No manga found with name: "+name)
		}

		replyKeyboard := [][]tb.ReplyButton{}
		replyKeys := []tb.ReplyButton{}

		for _, manga := range res.Suggestions {
			fmt.Println("The manga result is: ", manga.Data, manga.Value)

			replyBtn := tb.ReplyButton{
				Text: manga.Value,
			}

			bot.Handle(&replyBtn, func(bm *tb.Message) {
				mangaURL := fmt.Sprintf(feed.ViewManga(), manga.Data)
				bot.Send(bm.Sender, mangaURL)
			})

			replyKeys = append(replyKeys, replyBtn)
		}

		replyKeyboard = append(replyKeyboard, replyKeys)

		fmt.Println("Keyboard: ", replyKeyboard)
		bot.Send(m.Sender, "These are the manga I found ", &tb.ReplyMarkup{
			ReplyKeyboard: replyKeyboard,
		})
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
