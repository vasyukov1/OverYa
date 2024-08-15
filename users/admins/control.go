package admins

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
	"strconv"
	"strings"
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

func HandleAdminRequests(bot *tgbotapi.BotAPI, message *tgbotapi.Message, chatID int64, db *database.DB) {
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

	//editAdminRequestList(bot, message, db)
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

	//editAdminRequestList(bot, callbackQuery.Message, db)
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

//

func HandleGetAdminsInfo(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	adminsList := db.GetAdmins()
	if len(adminsList) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Список админов пуст.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send message error to %v: %v", chatID, err)
		}
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for adminID := range adminsList {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("Admin ID: %d", adminID),
			fmt.Sprintf("get_admin_info_%d", adminID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, "Выберите ID админа для получения информации:")
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Send message error to %v: %v", chatID, err)
	}
}

func HandleAdminInfoCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *database.DB) {
	data := strings.TrimPrefix(callbackQuery.Data, "get_admin_info_")
	adminID, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		log.Printf("Error parsing admin ID: %v\n", err)
		return
	}

	adminInfo, err := db.GetAdminInfo(adminID)
	if err != nil {
		log.Printf("Failed to get admin info: %v", err)
	}

	//adminsList := db.GetAdmins()

	//adminInfo, exists := adminsList[adminID]
	//if !exists {
	//	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "Информация об админе не найдена.")
	//	if _, err := bot.Send(msg); err != nil {
	//		log.Printf("Send message error to %v: %v", callbackQuery.Message.Chat.ID, err)
	//	}
	//	return
	//}

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, fmt.Sprintf("Информация об админе %d:\n%s", adminID, adminInfo))
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Send message error to %v: %v", callbackQuery.Message.Chat.ID, err)
	}
}
