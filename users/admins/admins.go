package admins

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
)

func AddAdmin(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	db.AddAdmin(chatID)
	log.Printf("Added admin %v", chatID)
	msg := tgbotapi.NewMessage(chatID, "You are now an admin!")
	bot.Send(msg)
}

func SendAdminRequest(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64, db *database.DB, chat *tgbotapi.Chat) {
	firstName := chat.FirstName
	lastName := chat.LastName
	userName := chat.UserName
	if err := db.AddAdminRequest(chatID, firstName, lastName, userName); err != nil {
		log.Printf("Send admin request error: %v", err)
		msg := tgbotapi.NewMessage(chatID, "We have problem with your admin request, sorry")
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(mainAdmin, "")
		msg.Text = fmt.Sprintf("You have new admin request! \nAll /admin_requests: %v", db.CountAdminRequest())
		bot.Send(msg)
	}
}

func DeleteAdmin(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64, db *database.DB) bool {
	if !db.IsAdmin(chatID) {
		log.Printf("Can't delete %v, is not an admin", chatID)
		msg := tgbotapi.NewMessage(mainAdmin, "")
		msg.Text = "You are not an admin."
		bot.Send(msg)
		return false
	}
	db.DeleteAdmin(chatID)
	msg := tgbotapi.NewMessage(chatID, "You aren't now an admin(")
	bot.Send(msg)
	return true
}
