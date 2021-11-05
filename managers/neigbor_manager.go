package managers

import (
	"errors"
	"fmt"
	"main/database"
	"main/models"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func NewNeighborManager(repository database.ConcurrentNeighborRepository, tg_bot tgbotapi.BotAPI) *NeighborManager {
	return &NeighborManager{repo: &repository, bot: &tg_bot}
}

type NeighborManager struct {
	repo *database.ConcurrentNeighborRepository
	bot  *tgbotapi.BotAPI
}

func (nm NeighborManager) RegisterNeighbor(message tgbotapi.Message) {

	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	msg.ReplyToMessageID = message.MessageID

	neighbor, parseErr := parseNeigbor(message)
	if parseErr != nil {
		msg.Text = parseErr.Error()
		nm.bot.Send(msg)
		return
	}

	validateErr := validateNeighbor(neighbor)
	if validateErr != nil {
		msg.Text = validateErr.Error()
		nm.bot.Send(msg)
		return
	}

	nm.repo.Upsert(neighbor)
	msg.Text = "Successsfully added user " + neighbor.Name
	nm.bot.Send(msg)
}

func (nm NeighborManager) ShowList(message tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Error")
	msg.ReplyToMessageID = message.MessageID

	neighbors, err := nm.repo.GetAll()
	if err == nil {
		var list = ""

		for _, n := range neighbors {
			list = list + fmt.Sprintf("Имя %s, корпус: %d, секция: %d \n", n.Name, n.Building, n.Section)
		}
		msg.Text = list
	} else {
		msg.Text = err.Error()
	}

	nm.bot.Send(msg)
}

func (nm NeighborManager) Delete(message tgbotapi.Message) {

	msg := tgbotapi.NewMessage(message.Chat.ID, "Error: ")
	msg.ReplyToMessageID = message.MessageID

	err := nm.repo.Delete(message.From.UserName)
	if err == nil {
		msg.Text = "Successfully deleted records for username: " + message.From.UserName
	} else {
		msg.Text += err.Error()
	}

	nm.bot.Send(msg)
}

func (nm NeighborManager) Me(message tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Error: ")
	msg.ReplyToMessageID = message.MessageID

	neighbors, err := nm.repo.GetByTelegramName(message.From.UserName)
	if err == nil {
		var list = fmt.Sprintf("Список зарегистрированных на вашего телеграмм пользователя (%d): \n", len(neighbors))

		for _, n := range neighbors {
			list = list + fmt.Sprintf("Имя %s,очередь:%d корпус: %d, секция: %d \n", n.Name, n.Turn, n.Building, n.Section)
		}
		msg.Text = list
	} else {
		msg.Text += err.Error()
	}

	nm.bot.Send(msg)
}

func (nm NeighborManager) About(message tgbotapi.Message) {

	msg := tgbotapi.NewMessage(message.Chat.ID,
		`Здравствуйте это бот чата дольщиков Нового Лесснера
		Доступные команды: 
		/me - показать записи зарегистрированные с моего телеграмм-аккаунта
		/del - удалить записи зарегистрированные с моего телеграмм-аккаунта
		/reg {очередь} {корпус} {секция} {Имя}  - зарегистрировать свой корпус + секцию + Имя. 
			Имя - необязательный параметр(в случае его отсутсвтия будут использованные данные из телеграмма). 
			Например "/reg 1 1 4 Анатолий" или "/reg 1 1 4"
		/list - вывести список зарегистрированных
	`)
	msg.ReplyToMessageID = message.MessageID

	nm.bot.Send(msg)
}

func validateNeighbor(n models.Neighbor) error {
	if n.Turn > 2 || n.Turn < 1 {
		return errors.New("укажите правильный номер очереди 1 или 2")
	}

	if n.Turn == 1 && (n.Building < 1 || n.Building > 4) {
		return errors.New("укажите правильный номер корпуса для 1ой очереди 1-4")
	}

	if n.Turn == 2 && (n.Building < 1 || n.Building > 7) {
		return errors.New("укажите правильный номер корпуса для 2ой очереди 1-7")
	}

	if n.Section < 1 || n.Section > 20 {
		return errors.New("укажите правильный номер секции от 1 до 20")
	}
	return nil
}

func parseNeigbor(message tgbotapi.Message) (models.Neighbor, error) {
	var neighbor = models.Neighbor{
		TelegramFirstName: message.From.FirstName,
		TelegramLastName:  message.From.LastName,
		TelegramUserName:  message.From.UserName,
	}

	words := strings.Fields(message.Text)
	if len(words) < 4 {
		return neighbor, errors.New("wrong params amount")
	}

	turn, err := strconv.Atoi(words[1])
	if err != nil {
		return neighbor, errors.New("can't parse turn number")
	}

	neighbor.Turn = int64(turn)
	building, err := strconv.Atoi(words[2])
	if err != nil {
		return neighbor, errors.New("can't parse building number")
	}

	neighbor.Building = int64(building)

	section, err := strconv.Atoi(words[3])
	if err != nil {
		return neighbor, errors.New("can't parse section number")
	}

	neighbor.Section = int64(section)

	if len(words) == 5 {
		neighbor.Name = words[4]
	} else {
		neighbor.Name = message.From.FirstName + " " + message.From.LastName
	}

	return neighbor, nil
}
