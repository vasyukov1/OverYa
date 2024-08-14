package admins

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
	"strconv"
	"strings"
)

func HandleAdminRequests(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	if !db.IsAdmin(chatID) {
		msg := tgbotapi.NewMessage(chatID, "У вас нет прав главного администратора.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	requests, err := db.GetAdminRequests()
	if err != nil {
		log.Printf("Failed to get admin requests: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Не удалось получить заявки на админство.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	if len(requests) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Заявок на админство нет.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, request := range requests {
		log.Printf("Processing admin request: %d", request)
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("Request %d", request), fmt.Sprintf("request_admin_%d", request))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	msg := tgbotapi.NewMessage(chatID, "Admin Requests:")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

func HandleAdminRequestCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *database.DB, inProcessAdminReq *map[int64]bool) {
	chatID := callbackQuery.Message.Chat.ID
	data := callbackQuery.Data

	var requestID int64
	var err error

	if strings.HasPrefix(data, "request_admin_") {
		requestIDStr := strings.TrimPrefix(data, "request_admin_")
		requestID, err = strconv.ParseInt(requestIDStr, 10, 64)
		if err != nil {
			log.Printf("Failed to parse request ID: %v", err)
			return
		}

		log.Printf("Handling admin request: %d", requestID)
		userInfo, err := db.GetAdminRequestInfo(requestID)
		if err != nil {
			log.Printf("Failed to get admin request info: %v", err)
			return
		}

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Would you like to accept ADMIN request from ID: %d?\nInfo: %s", requestID, userInfo))
		button := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Повысить", fmt.Sprintf("accept_admin_%d", requestID)),
				tgbotapi.NewInlineKeyboardButtonData("Отказать", fmt.Sprintf("reject_admin_%d", requestID)),
			),
		)
		msg.ReplyMarkup = button
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}

	} else if strings.HasPrefix(data, "accept_admin_") {
		requestIDStr := strings.TrimPrefix(data, "accept_admin_")
		requestID, err = strconv.ParseInt(requestIDStr, 10, 64)
		if err != nil {
			log.Printf("Failed to parse request ID: %v", err)
			return
		}
		log.Printf("Accepting admin request: %d", requestID)
		db.AddAdmin(requestID)

		msg := tgbotapi.NewMessage(requestID, "Вас повысили, теперь вы админ!")
		msgAdmin := tgbotapi.NewMessage(chatID, fmt.Sprintf("%d повышен!", requestID))
		if _, err := bot.Send(msgAdmin); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", requestID, err)
		}

	} else if strings.HasPrefix(data, "reject_admin_") {
		requestIDStr := strings.TrimPrefix(data, "reject_admin_")
		requestID, err = strconv.ParseInt(requestIDStr, 10, 64)
		if err != nil {
			log.Printf("Failed to parse request ID: %v", err)
			return
		}
		log.Printf("Rejecting admin request: %d", requestID)

		msg := tgbotapi.NewMessage(requestID, "Вам отказано в повышении.")
		msgAdmin := tgbotapi.NewMessage(chatID, fmt.Sprintf("%d остался на прежнем уровне.", requestID))
		if _, err := bot.Send(msgAdmin); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", requestID, err)
		}
	}

	if errDB := db.DeleteAdminRequest(requestID); errDB != nil {
		log.Printf("Failed to delete admin request: %v", err)
	}

	editAdminRequestList(bot, callbackQuery.Message, db)
	(*inProcessAdminReq)[requestID] = false
}

func editAdminRequestList(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *database.DB) {
	requests, err := db.GetAdminRequests()
	if err != nil {
		log.Printf("Failed to get admin requests: %v", err)
		return
	}

	if len(requests) == 0 {
		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, "Заявок на админство нет.")
		if _, err := bot.Send(editMsg); err != nil {
			log.Printf("Failed to edit message: %v", err)
		}
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, request := range requests {
		log.Printf("Updating request list with: %d", request)
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", request), fmt.Sprintf("request_admin_%d", request))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	editMsg := tgbotapi.NewEditMessageReplyMarkup(message.Chat.ID, message.MessageID, keyboard)
	if _, err := bot.Send(editMsg); err != nil {
		log.Printf("Failed to edit message: %v", err)
	}
}
