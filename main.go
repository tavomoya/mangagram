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
	publicURL := os.Getenv("PUBLIC_URL")
	token := os.Getenv("TOKEN")

	if port == "" {
		port = "9000"
	}

	listen := fmt.Sprintf(":%s", port)

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

	// Available commands:

	bot.Handle("/manga", func(m *tb.Message) {
		fmt.Println("The message received is ", m.Text)

		name := strings.Replace(m.Text, "/manga ", "", 1)

		if name == "" {
			bot.Send(m.Sender, "<b>No manga name supplied</b>", tb.ModeHTML)
		}

		feed := actions.NewMangaInterface(2)

		res := feed.QueryManga(name)
		if res == nil {
			bot.Send(m.Sender, "No Manga found with your criteria")
		}

		msg := "These are the manga I found:\n"

		inlineKb := [][]tb.InlineButton{}

		for _, item := range res.Suggestions {
			fmt.Println("The manga iten: ", item)
			inlineBtn := []tb.InlineButton{
				tb.InlineButton{
					Text:   item.Value + " 📖",
					Unique: item.Data,
					URL:    fmt.Sprintf(feed.ViewManga(), item.Data),
				},
				tb.InlineButton{
					Text:   "Subscribe" + " 🔔",
					Unique: item.Data + "_sub",
				},
			}

			bot.Handle(&inlineBtn[1], func(btnCb *tb.Callback) {
				fmt.Println("Subscribing user: ", btnCb.Sender.FirstName)
				bot.Respond(btnCb, &tb.CallbackResponse{
					Text:      "Succesfully subscribed",
					ShowAlert: true,
				})
			})

			fmt.Println("Inline btns: ", inlineBtn)
			inlineKb = append(inlineKb, inlineBtn)
		}

		fmt.Println("Final message and keyboard: ", msg, inlineKb)
		bot.Send(m.Sender, msg, &tb.ReplyMarkup{
			InlineKeyboard: inlineKb,
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
