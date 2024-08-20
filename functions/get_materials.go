package functions

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
	"sort"
	"strconv"
	"strings"
)

func HandleGetSubjects(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, db *database.DB, page int) {
	const itemsPerPageFirstLast = 9
	const itemsPerPageMiddle = 8
	subjects := db.GetSubjects()
	if len(subjects) == 0 {
		msg := tgbotapi.NewMessage(chatID, "No subjects found.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send message error to %v: %v", chatID, err)
		}
		return
	}

	sort.Strings(subjects)
	var startIndex, endIndex int
	if page == 0 {
		startIndex = 0
		endIndex = itemsPerPageFirstLast
	} else {
		startIndex = itemsPerPageFirstLast + (page-1)*itemsPerPageMiddle
		endIndex = startIndex + itemsPerPageMiddle
	}

	if endIndex > len(subjects) {
		endIndex = len(subjects)
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, subject := range subjects[startIndex:endIndex] {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s", subject), fmt.Sprintf("subject_%s", subject))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}

	var navigationButtons []tgbotapi.InlineKeyboardButton
	if page > 0 {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("subjects_page_%d", page-1)))
	}
	if endIndex < len(subjects) {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("Дальше ➡️", fmt.Sprintf("subjects_page_%d", page+1)))
	}
	if len(navigationButtons) > 0 {
		buttons = append(buttons, navigationButtons)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	if update.CallbackQuery == nil {
		msg := tgbotapi.NewMessage(chatID, "Select a subject:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send message error to %v: %v", chatID, err)
		}
	} else {
		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a subject:")
		editMsg.ReplyMarkup = &keyboard
		if _, err := bot.Send(editMsg); err != nil {
			log.Printf("Edit message error to %v: %v", chatID, err)
		}
	}
}

func handleGetControlElements(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, db *database.DB, subject string, page int) {
	const itemsPerPageFirstLast = 8
	const itemsPerPageMiddle = 7
	controlElements := db.GetControlElements(subject)
	if len(controlElements) == 0 {
		msg := tgbotapi.NewMessage(chatID, "No control elements found.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send message error to %v: %v", chatID, err)
		}
		return
	}

	sort.Strings(controlElements)
	var startIndex, endIndex int
	if page == 0 {
		startIndex = 0
		endIndex = itemsPerPageFirstLast
	} else {
		startIndex = itemsPerPageFirstLast + (page-1)*itemsPerPageMiddle
		endIndex = startIndex + itemsPerPageMiddle
	}

	if endIndex > len(controlElements) {
		endIndex = len(controlElements)
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, controlElement := range controlElements[startIndex:endIndex] {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s", controlElement), fmt.Sprintf("control_%s", controlElement))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}

	var navigationButtons []tgbotapi.InlineKeyboardButton
	if page > 0 {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("controls_page_%d", page-1)))
	}
	if endIndex < len(controlElements) {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("Дальше ➡️", fmt.Sprintf("controls_page_%d", page+1)))
	}
	if len(navigationButtons) > 0 {
		buttons = append(buttons, navigationButtons)
	}
	navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_subjects"))
	buttons = append(buttons, navigationButtons)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a control element:")
	editMsg.ReplyMarkup = &keyboard
	if _, err := bot.Send(editMsg); err != nil {
		log.Printf("Edit message error to %v: %v", chatID, err)
	}
}

func handleGetElementNumbers(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, db *database.DB, subject, controlElement string, page int) {
	const itemsPerPageFirstLast = 8
	const itemsPerPageMiddle = 7
	elementNumbers := db.GetElementNumber(subject, controlElement)
	if len(elementNumbers) == 0 {
		msg := tgbotapi.NewMessage(chatID, "No element numbers found.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send message error to %v: %v", chatID, err)
		}
		return
	}

	sort.Ints(elementNumbers)
	var startIndex, endIndex int
	if page == 0 {
		startIndex = 0
		endIndex = itemsPerPageFirstLast
	} else {
		startIndex = itemsPerPageFirstLast + (page-1)*itemsPerPageMiddle
		endIndex = startIndex + itemsPerPageMiddle
	}

	if endIndex > len(elementNumbers) {
		endIndex = len(elementNumbers)
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, number := range elementNumbers[startIndex:endIndex] {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", number), fmt.Sprintf("number_%d", number))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}

	var navigationButtons []tgbotapi.InlineKeyboardButton
	if page > 0 {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("elements_page_%d", page-1)))
	}
	if endIndex < len(elementNumbers) {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("Дальше ➡️", fmt.Sprintf("elements_page_%d", page+1)))
	}
	navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_controls"))
	buttons = append(buttons, navigationButtons)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a number:")
	editMsg.ReplyMarkup = &keyboard
	if _, err := bot.Send(editMsg); err != nil {
		log.Printf("Edit message error to %v: %v", chatID, err)
	}
}

func HandleCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *database.DB, telegramChannel int64, isBroadcastMode *map[int64]bool) {
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackData := update.CallbackQuery.Data

	if db.IsAdmin(chatID) {
		switch callbackData {
		case "yes_description":
			msg := tgbotapi.NewMessage(chatID, "Please provide the description")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Failed to send message to %v: %v", chatID, err)
			}
			isDescriptionMode[chatID] = true
		case "no_description":
			sendBroadcast(bot, chatID, db, telegramChannel, isBroadcastMode)
		}
	}

	if strings.HasPrefix(callbackData, "subjects_page_") {
		pageStr := strings.TrimPrefix(callbackData, "subjects_page_")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid page number: %v", err)
			return
		}
		HandleGetSubjects(bot, update, chatID, db, page)
		answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(answer); err != nil {
			log.Printf("Error sending callback response: %v", err)
		}

	} else if strings.HasPrefix(callbackData, "subject_") {
		subject := strings.TrimPrefix(callbackData, "subject_")
		userSubject[chatID] = subject
		log.Printf("User %v choose subject: %v", chatID, subject)
		handleGetControlElements(bot, update, chatID, db, subject, 0)

	} else if strings.HasPrefix(callbackData, "controls_page_") {
		subject := userSubject[chatID]
		pageStr := strings.TrimPrefix(callbackData, "controls_page_")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid page number: %v", err)
			return
		}
		handleGetControlElements(bot, update, chatID, db, subject, page)
		answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(answer); err != nil {
			log.Printf("Error sending callback response: %v", err)
		}

	} else if strings.HasPrefix(callbackData, "control_") {
		controlElement := strings.TrimPrefix(callbackData, "control_")
		userControlElement[chatID] = controlElement
		log.Printf("User %v choose control element: %v", chatID, controlElement)
		handleGetElementNumbers(bot, update, chatID, db, userSubject[chatID], controlElement, 0)

	} else if strings.HasPrefix(callbackData, "elements_page_") {
		controlElement := userControlElement[chatID]
		pageStr := strings.TrimPrefix(callbackData, "elements_page_")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid page number: %v", err)
			return
		}
		handleGetElementNumbers(bot, update, chatID, db, userSubject[chatID], controlElement, page)
		answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(answer); err != nil {
			log.Printf("Error sending callback response: %v", err)
		}

	} else if strings.HasPrefix(callbackData, "number_") {
		numberStr := strings.TrimPrefix(callbackData, "number_")
		number, err := strconv.Atoi(numberStr)
		if err != nil {
			log.Println("Invalid number:", numberStr)
			return
		}
		log.Printf("User %v choose element number: %v", chatID, number)
		subject := userSubject[chatID]
		controlElement := userControlElement[chatID]
		SendMaterial(bot, chatID, db, subject, controlElement, number)

	} else if callbackData == "back_to_subjects" {
		HandleGetSubjects(bot, update, chatID, db, 0)

	} else if callbackData == "back_to_controls" {
		subject := userSubject[chatID]
		handleGetControlElements(bot, update, chatID, db, subject, 0)

	}

	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Callback error: %v", err)
	}
}
