package main

import (
	"log"
	"os"

	"main/database"

	"main/managers"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
)

func main() {

	//Databse
	//ToDo: make const -> move to env manager

	var CONNECTION_STRING = os.Getenv("NL_BOT_CS")
	var BOT_ACCCESS_TOCKEN = os.Getenv("NL_BOT_AT")

	conn, _ := sqlx.Connect("pgx", CONNECTION_STRING)
	neighborRepository := database.NewConcurrentNeighborRepository(*conn)

	//Bot
	bot, err := tgbotapi.NewBotAPI(BOT_ACCCESS_TOCKEN)
	if err != nil {
		log.Panic(err)
	}

	//Command handler
	neighborManager := managers.NewNeighborManager(*neighborRepository, *bot)

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || len(update.Message.Text) == 0 {
			continue
		}

		messageText := update.Message.Text
		words := strings.Fields(messageText)
		command := words[0]
		log.Printf("Recieved command: %s , full message: %s", command, update.Message.Text)

		switch command {
		case "/start":
			neighborManager.About(*update.Message)

		case "/list":
			neighborManager.ShowList(*update.Message)

		case "/reg":
			neighborManager.RegisterNeighbor(*update.Message)

		case "/del":
			neighborManager.Delete(*update.Message)

		case "/me":
			neighborManager.Me(*update.Message)

		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command, use /start to show available")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	}
}
