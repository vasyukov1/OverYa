package users

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
)

func AdminRequest(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64, db *database.DB) {
	if db.IsAdmin(chatID) {
		msg := tgbotapi.NewMessage(chatID, "You are already an admin")
		bot.Send(msg)
		return
	}
	msg := tgbotapi.NewMessage(mainAdmin, "")
	msg.Text = fmt.Sprintf("User %v wants to be an admin", chatID)
	bot.Send(msg)
}

// AnswerToAdminRequest This function is work when main admin
// click on "answer to admin request"
func AnswerToAdminRequest(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64) {
	msg := tgbotapi.NewMessage(mainAdmin, fmt.Sprintf("Do you want to add %v as an admin?", chatID))
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", "accept"),
			tgbotapi.NewInlineKeyboardButtonData("No", "decline"),
		),
	)
	msg.ReplyMarkup = buttons
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
	// if answer is "accept" go to function "Add admin"
}

func ShowAdminRequest(bot *tgbotapi.BotAPI, mainAdmin int64) {
	// Показ заявок на админа.
}

func AddAdmin(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	db.AddAdmin(chatID)
	log.Printf("Added admin %v", chatID)
	msg := tgbotapi.NewMessage(chatID, "You are now an admin!")
	bot.Send(msg)
}

func DeleteAdmin(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64, db *database.DB) {
	if !db.IsAdmin(chatID) {
		log.Printf("Can't delete %v, is not an admin", chatID)
		msg := tgbotapi.NewMessage(mainAdmin, "")
		msg.Text = "You are not an admin."
		bot.Send(msg)
		return
	}
	db.DeleteAdmin(chatID)
	msg := tgbotapi.NewMessage(chatID, "You aren't now an admin(")
	bot.Send(msg)
}
