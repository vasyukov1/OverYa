package functions

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
)

func News(bot *tgbotapi.BotAPI, adminID int64, message string, db *database.DB) {
	subscribers := db.GetSubscribers()
	isSend := true
	for chatID := range subscribers {
		msg := tgbotapi.NewMessage(chatID, "")
		msg.Text = fmt.Sprintf("NEWS\n\n" + message)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
			isSend = false
		}
	}
	msg := tgbotapi.NewMessage(adminID, "")
	if isSend {
		msg.Text = "Новость отправлена"
	} else {
		msg.Text = "Произошла ошибка при отправке новости"
	}

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", adminID, err)
	}
}
