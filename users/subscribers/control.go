package subscribers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
	"strconv"
	"strings"
)

func HandleSubscriberRequests(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	if !db.IsAdmin(chatID) {
		msg := tgbotapi.NewMessage(chatID, "У вас нет прав администратора.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	requests, err := db.GetSubscriberRequests()
	if err != nil {
		log.Printf("Failed to get subscriber requests: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Не удалось получить заявки на подписку.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	if len(requests) == 0 {
		msg := tgbotapi.NewMessage(chatID, "There are no requests.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, request := range requests {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", request), fmt.Sprintf("request_subscriber_%d", request))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	msg := tgbotapi.NewMessage(chatID, "Subscriber Requests:")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

func HandleRequestCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *database.DB, inProcessSubReq *map[int64]bool) {
	chatID := callbackQuery.Message.Chat.ID
	data := callbackQuery.Data

	if strings.HasPrefix(data, "request_subscriber_") {
		requestIDStr := strings.TrimPrefix(data, "request_subscriber_")
		requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
		if err != nil {
			log.Printf("Failed to parse request ID: %v", err)
			return
		}

		userInfo, err := db.GetSubscriberRequestInfo(requestID)
		if err != nil {
			log.Printf("Failed to get subscriber request info: %v", err)
			return
		}

		msg := tgbotapi.NewMessage(chatID, "")
		msg.Text = fmt.Sprintf("Would you like to accept request from ID: %d?\nInfo: %s", requestID, userInfo)

		button := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Принять", fmt.Sprintf("accept_subscriber_%d", requestID)),
				tgbotapi.NewInlineKeyboardButtonData("Отклонить", fmt.Sprintf("reject_subscriber_%d", requestID)),
			),
		)
		msg.ReplyMarkup = button
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	}

	var requestID int64
	var err error

	if strings.HasPrefix(data, "accept_subscriber_") {
		requestIDStr := strings.TrimPrefix(data, "accept_subscriber_")
		requestID, err = strconv.ParseInt(requestIDStr, 10, 64)
		if err != nil {
			log.Printf("Failed to parse request ID: %v", err)
			return
		}
		db.AddSubscriber(requestID)
		msg := tgbotapi.NewMessage(requestID, "Ваша заявка принята, теперь вы подписчик!")
		msgAdmin := tgbotapi.NewMessage(chatID, fmt.Sprintf("Заявка от %d принята.", requestID))
		if _, err := bot.Send(msgAdmin); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", requestID, err)
		}
	} else if strings.HasPrefix(data, "reject_subscriber_") {
		requestIDStr := strings.TrimPrefix(data, "reject_subscriber_")
		requestID, err = strconv.ParseInt(requestIDStr, 10, 64)
		if err != nil {
			log.Printf("Failed to parse request ID: %v", err)
			return
		}
		msg := tgbotapi.NewMessage(requestID, "Ваша заявка отклонена.")
		msgAdmin := tgbotapi.NewMessage(chatID, fmt.Sprintf("Заявка от %d отклонена.", requestID))
		if _, err := bot.Send(msgAdmin); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", requestID, err)
		}
	}
	if errDB := db.DeleteSubscriberRequest(requestID); errDB != nil {
		log.Printf("Failed to delete subscriber request: %v", err)
	}
	editRequestList(bot, callbackQuery.Message, db)
	(*inProcessSubReq)[requestID] = false
}

func editRequestList(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *database.DB) {
	requests, err := db.GetSubscriberRequests()
	if err != nil {
		log.Printf("Failed to get subscriber requests: %v", err)
		return
	}
	if len(requests) == 0 {
		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, "Заявок нет.")
		if _, err := bot.Send(editMsg); err != nil {
			log.Printf("Failed to edit message: %v", err)
		}
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, request := range requests {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", request), fmt.Sprintf("request_subscriber_%d", request))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	if message.ReplyMarkup != nil && message.ReplyMarkup.InlineKeyboard != nil {
		if len(message.ReplyMarkup.InlineKeyboard) == len(keyboard.InlineKeyboard) {
			same := true
			for i, row := range message.ReplyMarkup.InlineKeyboard {
				if len(row) != len(keyboard.InlineKeyboard[i]) {
					same = false
					break
				}
				for j, button := range row {
					if button.Text != keyboard.InlineKeyboard[i][j].Text || button.CallbackData == nil || *button.CallbackData != *keyboard.InlineKeyboard[i][j].CallbackData {
						same = false
						break
					}
				}
				if !same {
					break
				}
			}
			if same {
				return
			}
		}
	}

	editMsg := tgbotapi.NewEditMessageReplyMarkup(message.Chat.ID, message.MessageID, keyboard)
	if _, err := bot.Send(editMsg); err != nil {
		log.Printf("Failed to edit message: %v", err)
	}
}
