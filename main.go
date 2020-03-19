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
			bot.Send(m.Sender, "No manga name supplied")
		}

		feed := actions.NewMangaInterface(2)

		res := feed.QueryManga(name)
		if res == nil {
			bot.Send(m.Sender, "No manga found with name: "+name)
		}

		msg := "<b>This is what I found:</b>\n\n"
		mangas := make([]string, len(res.Suggestions))

		for i, manga := range res.Suggestions {
			fmt.Println("The manga result is: ", manga.Data, manga.Value)

			t := fmt.Sprintf("%s - <a href='%s'>%s</a>", manga.Value, fmt.Sprintf(feed.ViewManga(), manga.Data), manga.Value)

			mangas[i] = t
		}

		msg = fmt.Sprintf("%s%s", msg, strings.Join(mangas, "\n"))

		fmt.Println("Msg: ", msg)
		_, err := bot.Send(m.Sender, msg, tb.ModeHTML, tb.NoPreview)
		if err != nil {
			log.Println("There was an error sending the message: ", err.Error())
		}
	})

	bot.Handle("/subscribe", func(m *tb.Message) {
		fmt.Println("This is the subscribe command", m.Text)
		mangaQuery := strings.Replace(m.Text, "/subscribe ", "", 1)

		if mangaQuery == "" {
			msg := `
					This command allows you to subscribe to manga alerts: \n

					/subscribe Bleach
			`
			bot.Send(m.Sender, msg)
		}

		feed := actions.NewMangaInterface(2)

		res := feed.QueryManga(mangaQuery)
		if res == nil {
			bot.Send(m.Sender, "No Manga found with your criteria")
		}

		if len(res.Suggestions) == 1 {
			fmt.Println("Subscribing user: ", m.Sender.FirstName)
			bot.Send(m.Sender, "Succesfully subscribed to "+mangaQuery)
		}

		msg := "<b>Select the manga you want to subscribe to:<b>\n"

		inlineKb := [][]tb.InlineButton{}
		inlineKeys := []tb.InlineButton{}

		for _, item := range res.Suggestions {

			inlineBtn := tb.InlineButton{
				Text:   item.Value,
				Unique: item.Data,
			}

			bot.Handle(&inlineBtn, func(btnCb *tb.Callback) {
				fmt.Println("Subscribing user: ", btnCb.Sender.FirstName)
				bot.Respond(btnCb, &tb.CallbackResponse{
					Text: "Succesfully subscribed to " + mangaQuery,
				})
			})

			inlineKeys = append(inlineKeys, inlineBtn)
		}

		inlineKb = append(inlineKb, inlineKeys)

		bot.Send(m.Sender, msg, &tb.ReplyMarkup{
			InlineKeyboard: inlineKb,
		}, tb.ModeHTML)
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
