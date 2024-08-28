package subscribers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
	"strconv"
	"strings"
)

func AddSubscriber(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	db.AddSubscriber(chatID)
	log.Printf("Added subscriber %v", chatID)
	msg := tgbotapi.NewMessage(chatID, "You are now a subscriber!")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func SendSubscribeRequest(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64, db *database.DB, chat *tgbotapi.Chat) {
	firstName := chat.FirstName
	lastName := chat.LastName
	userName := chat.UserName
	if err := db.AddSubscriberRequest(chatID, firstName, lastName, userName); err != nil {
		log.Printf("Send subscribe request error: %v", err)
		msg := tgbotapi.NewMessage(chatID, "We have problem with your request, sorry")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(mainAdmin, "")
		msg.Text = fmt.Sprintf("You have new subscriber request! \nAll /requests_subscriber: %v", db.CountSubscriberRequest())
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

func DeleteSubscriber(bot *tgbotapi.BotAPI, chatID int64, mainAdmin int64, db *database.DB) bool {
	if !db.IsSubscriber(chatID) {
		log.Printf("Can't delete %v, is not a subscriber", chatID)
		msg := tgbotapi.NewMessage(mainAdmin, "")
		msg.Text = "It is not a subscriber."
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return false
	}
	if db.IsAdmin(chatID) {
		db.DeleteAdmin(chatID)
		msg := tgbotapi.NewMessage(chatID, "*You aren't now an admin*")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
	db.DeleteSubscriber(chatID)
	msg := tgbotapi.NewMessage(chatID, "*You aren't now a subscriber(*")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
	return true
}

func GetSubscribers(bot *tgbotapi.BotAPI, update tgbotapi.Update, adminID int64, db *database.DB, page int) {
	const itemsPerPageFirstLast = 9
	const itemsPerPageMiddle = 8

	if !db.IsAdmin(adminID) {
		msg := tgbotapi.NewMessage(adminID, "У вас нет прав администратора.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", adminID, err)
		}
		return
	}

	subscribers := db.GetSubscribers()
	if len(subscribers) == 0 {
		msg := tgbotapi.NewMessage(adminID, "Подписчиков нет.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", adminID, err)
		}
		return
	}
	var subscriberIDs []int64
	for id := range subscribers {
		subscriberIDs = append(subscriberIDs, id)
	}

	var startIndex, endIndex int
	if page == 0 {
		startIndex = 0
		endIndex = itemsPerPageFirstLast
	} else {
		startIndex = itemsPerPageFirstLast + (page-1)*itemsPerPageMiddle
		endIndex = startIndex + itemsPerPageMiddle
	}

	if endIndex > len(subscribers) {
		endIndex = len(subscribers)
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, subscriber := range subscriberIDs[startIndex:endIndex] {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", subscriber), fmt.Sprintf("subscriber_%d", subscriber))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}

	var navigationButtons []tgbotapi.InlineKeyboardButton
	if page > 0 {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("page_subscribers_%d", page-1)))
	}
	if endIndex < len(subscribers) {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("Дальше ➡️", fmt.Sprintf("page_subscribers_%d", page+1)))
	}
	if len(navigationButtons) > 0 {
		buttons = append(buttons, navigationButtons)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	if update.CallbackQuery == nil {
		msg := tgbotapi.NewMessage(adminID, "Список подписчиков:")
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

func HandleSubscriberRequests(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, db *database.DB, page int) {
	const itemsPerPageFirstLast = 9
	const itemsPerPageMiddle = 8

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
		msg := tgbotapi.NewMessage(chatID, "Заявок на подписку нет.")
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
			fmt.Sprintf("%d", request), fmt.Sprintf("request_subscriber_%d", request))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}

	var navigationButtons []tgbotapi.InlineKeyboardButton
	if page > 0 {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("page_subscriber_requests%d", page-1)))
	}
	if endIndex < len(requests) {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("Дальше ➡️", fmt.Sprintf("page_subscriber_requests%d", page+1)))
	}
	if len(navigationButtons) > 0 {
		buttons = append(buttons, navigationButtons)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	if update.CallbackQuery == nil {
		msg := tgbotapi.NewMessage(chatID, "Заявки на подписку:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	} else {
		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Заявки на подписку:")
		editMsg.ReplyMarkup = &keyboard
		if _, err := bot.Send(editMsg); err != nil {
			log.Printf("Failed to edit message for %v: %v", chatID, err)
		}
	}
}

func HandleRequestCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *database.DB, inProcessSubReq *map[int64]bool) {
	chatID := callbackQuery.Message.Chat.ID
	data := callbackQuery.Data

	var requestID int64
	var err error

	if strings.HasPrefix(data, "request_subscriber_") {
		requestIDStr := strings.TrimPrefix(data, "request_subscriber_")
		requestID, err = strconv.ParseInt(requestIDStr, 10, 64)
		if err != nil {
			log.Printf("Failed to parse request ID: %v", err)
			return
		}

		log.Printf("Handling subscriber request: %d", requestID)
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

	if strings.HasPrefix(data, "accept_subscriber_") {
		requestIDStr := strings.TrimPrefix(data, "accept_subscriber_")
		requestID, err = strconv.ParseInt(requestIDStr, 10, 64)
		if err != nil {
			log.Printf("Failed to parse request ID: %v", err)
			return
		}
		log.Printf("Accepting subscriber request: %d", requestID)
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
		log.Printf("Rejecting subscriber request: %d", requestID)
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
	//editRequestList(bot, callbackQuery.Message, db)
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
