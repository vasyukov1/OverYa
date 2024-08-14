package users

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
)

func AddSubscriber(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	db.AddSubscriber(chatID)
	log.Printf("Added subscriber %v", chatID)
	msg := tgbotapi.NewMessage(chatID, "You are now a subscriber!")
	bot.Send(msg)
}

func SendSubscribeRequest(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64, db *database.DB, chat *tgbotapi.Chat) {
	firstName := chat.FirstName
	lastName := chat.LastName
	userName := chat.UserName
	if err := db.AddSubscriberRequest(chatID, firstName, lastName, userName); err != nil {
		log.Printf("Send subscribe request error: %v", err)
		msg := tgbotapi.NewMessage(chatID, "We have problem with your request, sorry")
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(mainAdmin, "")
		msg.Text = fmt.Sprintf("You have new subscriber request! \nAll /requests: %v", db.CountSubscriberRequest())
		bot.Send(msg)
	}
}

func DeleteSubscriber(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64, db *database.DB) bool {
	if !db.IsSubscriber(chatID) {
		log.Printf("Can't delete %v, is not a subscriber", chatID)
		msg := tgbotapi.NewMessage(mainAdmin, "")
		msg.Text = "It is not a subscriber."
		bot.Send(msg)
		return false
	}
	db.DeleteSubscriber(chatID)
	msg := tgbotapi.NewMessage(chatID, "**You aren't now a subscriber(**")
	bot.Send(msg)
	return true
}
