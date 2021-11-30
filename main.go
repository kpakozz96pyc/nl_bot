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

	bot.Debug = false

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
		var sendedMsg tgbotapi.Message
		var err error

		if update.Message.Chat.IsPrivate() {
			switch command {
			case "/nl_help":
				msg = neighborManager.About(*update.Message)
				msg.ReplyToMessageID = update.Message.MessageID
				sendedMsg, err = bot.Send(msg)

			case "/nl_list":
				msg = neighborManager.ShowList(*update.Message)
				msg.ReplyToMessageID = update.Message.MessageID
				sendedMsg, err = bot.Send(msg)

			case "/nl_reg":
				msg = neighborManager.RegisterNeighbor(*update.Message)
				msg.ReplyToMessageID = update.Message.MessageID
				sendedMsg, err = bot.Send(msg)

			case "/nl_del":
				msg = neighborManager.Delete(*update.Message)
				msg.ReplyToMessageID = update.Message.MessageID
				sendedMsg, err = bot.Send(msg)

			case "/nl_me":
				msg = neighborManager.Me(*update.Message)
				msg.ReplyToMessageID = update.Message.MessageID
				sendedMsg, err = bot.Send(msg)
			default:
				if update.Message.Chat.IsPrivate() {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command, use /nl_help to show available")
					sendedMsg, err = bot.Send(msg)
				}
			}

			//Delete bot messages if not private chat
			if (err == nil && sendedMsg != tgbotapi.Message{}) {
				if !update.Message.Chat.IsPrivate() {
					time.AfterFunc(time.Duration(15)*time.Second, func() {
						bot.DeleteMessage(tgbotapi.DeleteMessageConfig{ChatID: msg.ChatID, MessageID: sendedMsg.MessageID})
					})
				}
			}
		} else {
			switch command {
			case "/nl_help", "/nl_list", "/nl_reg", "/nl_del", "/nl_me":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Напишите боту в личку, и он вам ответит @nl_neighbor_bot")
				sendedMsg, _ = bot.Send(msg)
				time.AfterFunc(time.Duration(15)*time.Second, func() {
					bot.DeleteMessage(tgbotapi.DeleteMessageConfig{ChatID: msg.ChatID, MessageID: sendedMsg.MessageID})
					bot.DeleteMessage(tgbotapi.DeleteMessageConfig{ChatID: msg.ChatID, MessageID: update.Message.MessageID})
				})
			}
		}

	}
}
