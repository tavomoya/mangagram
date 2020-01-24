package main

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"os"
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
		bot.Send(m.Sender, "Tu mensaje fue: "+m.Text)
	})

}
