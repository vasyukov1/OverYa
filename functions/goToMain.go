package functions

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
)

func GoToMain(chatID int64, db *database.DB, bot *tgbotapi.BotAPI) {
	IsBroadcastMode[chatID] = false
	MaterialStep[chatID] = ""
	InProcessSubReq[chatID] = false
	InProcessAdminReq[chatID] = false
	IsDeleteSubscriberMode = false
	IsDeleteAdminMode = false
	IsDeleteMaterialMode = false
	IsDeleteSubjectMode = false

	msg := tgbotapi.NewMessage(chatID, "Hello, HSE Student!")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending message to %v: %v", chatID, err)
	}
}
