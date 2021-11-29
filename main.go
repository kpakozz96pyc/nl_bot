package main

import (
	"log"
	"os"
	"time"

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

	if len(CONNECTION_STRING) == 0 || len(BOT_ACCCESS_TOCKEN) == 0 {
		log.Panic("Set env variables for db and bot_token!")
	}

	conn, err := sqlx.Connect("pgx", CONNECTION_STRING)
	neighborRepository := database.NewConcurrentNeighborRepository(*conn)
	if err != nil {
		log.Panic(err)
	}

	//Bot
	bot, err := tgbotapi.NewBotAPI(BOT_ACCCESS_TOCKEN)
	if err != nil {
		log.Panic(err)
	}

	//Command handler
	neighborManager := managers.NewNeighborManager(*neighborRepository)

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
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command, use /nl_help to show available")

		switch command {
		case "/nl_help":
			msg = neighborManager.About(*update.Message)

		case "/nl_list":
			msg = neighborManager.ShowList(*update.Message)

		case "/nl_reg":
			msg = neighborManager.RegisterNeighbor(*update.Message)

		case "/nl_del":
			msg = neighborManager.Delete(*update.Message)

		case "/nl_me":
			msg = neighborManager.Me(*update.Message)

		default:
			msg.ReplyToMessageID = update.Message.MessageID
		}

		sendedMsg, err := bot.Send(msg)

		if err == nil {
			if !update.Message.Chat.IsPrivate() {
				time.AfterFunc(time.Duration(15)*time.Second, func() {
					bot.DeleteMessage(tgbotapi.DeleteMessageConfig{ChatID: msg.ChatID, MessageID: sendedMsg.MessageID})
				})
			}
		}

	}
}
