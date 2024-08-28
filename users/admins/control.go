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
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Send admin request error: %v", err)
	}
}

func SendAdminRequest(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64, db *database.DB, chat *tgbotapi.Chat) {
	firstName := chat.FirstName
	lastName := chat.LastName
	userName := chat.UserName
	if err := db.AddAdminRequest(chatID, firstName, lastName, userName); err != nil {
		log.Printf("Send admin request error: %v", err)
		msg := tgbotapi.NewMessage(chatID, "We have problem with your admin request, sorry")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send admin request error: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(mainAdmin, "")
		msg.Text = fmt.Sprintf("You have new admin request! \nAll /requests_admin: %v", db.CountAdminRequest())
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send admin request error: %v", err)
		}
	}
}

func DeleteAdmin(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64, db *database.DB) bool {
	if !db.IsAdmin(chatID) {
		log.Printf("Can't delete %v, is not an admin", chatID)
		msg := tgbotapi.NewMessage(mainAdmin, "")
		msg.Text = "You are not an admin."
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send admin request error: %v", err)
		}
		return false
	}
	db.DeleteAdmin(chatID)
	msg := tgbotapi.NewMessage(chatID, "You aren't now an admin(")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Send admin request error: %v", err)
	}
	return true
}

func HandleAdminRequests(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, db *database.DB, page int) {
	const itemsPerPageFirstLast = 9
	const itemsPerPageMiddle = 8

	if !db.IsAdmin(chatID) {
		msg := tgbotapi.NewMessage(chatID, "У вас нет прав администратора.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	requests, err := db.GetAdminRequests()
	if err != nil {
		log.Printf("Failed to get admin requests: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Не удалось получить заявки на админа.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	if len(requests) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Заявок на админа нет.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	var startIndex, endIndex int
	if page == 0 {
		startIndex = 0
		endIndex = itemsPerPageFirstLast
	} else {
		startIndex = itemsPerPageFirstLast + (page-1)*itemsPerPageMiddle
		endIndex = startIndex + itemsPerPageMiddle
	}

	if endIndex > len(requests) {
		endIndex = len(requests)
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, request := range requests[startIndex:endIndex] {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", request), fmt.Sprintf("request_admin_%d", request))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}

	var navigationButtons []tgbotapi.InlineKeyboardButton
	if page > 0 {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("page_admin_request_%d", page-1)))
	}
	if endIndex < len(requests) {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("Дальше ➡️", fmt.Sprintf("page_admin_request_%d", page+1)))
	}
	if len(navigationButtons) > 0 {
		buttons = append(buttons, navigationButtons)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	if update.CallbackQuery == nil {
		msg := tgbotapi.NewMessage(chatID, "Заявки на админа:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	} else {
		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Заявки на админа:")
		editMsg.ReplyMarkup = &keyboard
		if _, err := bot.Send(editMsg); err != nil {
			log.Printf("Failed to edit message for %v: %v", chatID, err)
		}
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

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, fmt.Sprintf("Информация об админе %d:\n%s", adminID, adminInfo))
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Send message error to %v: %v", callbackQuery.Message.Chat.ID, err)
	}
}

func GetAdmins(bot *tgbotapi.BotAPI, update tgbotapi.Update, adminID int64, db *database.DB, page int) {
	const itemsPerPageFirstLast = 9
	const itemsPerPageMiddle = 8

	if !db.IsAdmin(adminID) {
		msg := tgbotapi.NewMessage(adminID, "У вас нет прав администратора.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", adminID, err)
		}
		return
	}

	admins := db.GetAdmins()
	if len(admins) == 0 {
		msg := tgbotapi.NewMessage(adminID, "Админов нет.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", adminID, err)
		}
		return
	}
	var adminIDs []int64
	for id := range admins {
		adminIDs = append(adminIDs, id)
	}

	var startIndex, endIndex int
	if page == 0 {
		startIndex = 0
		endIndex = itemsPerPageFirstLast
	} else {
		startIndex = itemsPerPageFirstLast + (page-1)*itemsPerPageMiddle
		endIndex = startIndex + itemsPerPageMiddle
	}

	if endIndex > len(admins) {
		endIndex = len(admins)
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, admin := range adminIDs[startIndex:endIndex] {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", admin), fmt.Sprintf("admin_%d", admin))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}

	var navigationButtons []tgbotapi.InlineKeyboardButton
	if page > 0 {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("page_admins_%d", page-1)))
	}
	if endIndex < len(admins) {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("Дальше ➡️", fmt.Sprintf("page_admins_%d", page+1)))
	}
	if len(navigationButtons) > 0 {
		buttons = append(buttons, navigationButtons)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	if update.CallbackQuery == nil {
		msg := tgbotapi.NewMessage(adminID, "Список админов:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", adminID, err)
		}
	} else {
		editMsg := tgbotapi.NewEditMessageText(adminID, update.CallbackQuery.Message.MessageID, "Список подписчиков:")
		editMsg.ReplyMarkup = &keyboard
		if _, err := bot.Send(editMsg); err != nil {
			log.Printf("Failed to edit message for %v: %v", adminID, err)
		}
	}
}
