package functions

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"github.com/vasyukov1/Overbot/users/admins"
	"github.com/vasyukov1/Overbot/users/subscribers"
	"log"
	"strconv"
	"strings"
)

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

	if strings.HasPrefix(callbackData, "page_subjects_") {
		pageStr := strings.TrimPrefix(callbackData, "page_subjects_")
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
	} else if strings.HasPrefix(callbackData, "page_controls_") {
		subject := userSubject[chatID]
		pageStr := strings.TrimPrefix(callbackData, "page_controls_")
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
	} else if strings.HasPrefix(callbackData, "page_elements_") {
		controlElement := userControlElement[chatID]
		pageStr := strings.TrimPrefix(callbackData, "page_elements_")
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
	} else if strings.HasPrefix(callbackData, "page_subscribers_") {
		pageStr := strings.TrimPrefix(callbackData, "page_subscribers_")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid subscriber page number: %v", err)
			return
		}
		subscribers.GetSubscribers(bot, update, chatID, db, page)
		answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(answer); err != nil {
			log.Printf("Error sending callback response: %v", err)
		}
	} else if strings.HasPrefix(callbackData, "page_admins_") {
		pageStr := strings.TrimPrefix(callbackData, "page_admins_")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid subscriber page number: %v", err)
			return
		}
		admins.GetAdmins(bot, update, chatID, db, page)
		answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(answer); err != nil {
			log.Printf("Error sending callback response: %v", err)
		}
	} else if strings.HasPrefix(callbackData, "page_subscriber_requests") {
		pageStr := strings.TrimPrefix(callbackData, "page_subscriber_requests")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid subscriber page number: %v", err)
			return
		}
		subscribers.HandleSubscriberRequests(bot, update, chatID, db, page)
		answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(answer); err != nil {
			log.Printf("Error sending callback response: %v", err)
		}
	} else if strings.HasPrefix(callbackData, "page_admin_request_") {
		pageStr := strings.TrimPrefix(callbackData, "page_admin_request_")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Printf("Invalid subscriber page number: %v", err)
			return
		}
		admins.HandleAdminRequests(bot, update, chatID, db, page)
		answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(answer); err != nil {
			log.Printf("Error sending callback response: %v", err)
		}
	} else if callbackData == "back_to_subjects" {
		HandleGetSubjects(bot, update, chatID, db, 0)

	} else if callbackData == "back_to_controls" {
		subject := userSubject[chatID]
		handleGetControlElements(bot, update, chatID, db, subject, 0)
	} else if strings.HasPrefix(callbackData, "subject_") {
		subject := strings.TrimPrefix(callbackData, "subject_")
		userSubject[chatID] = subject
		log.Printf("User %v choose subject: %v", chatID, subject)
		handleGetControlElements(bot, update, chatID, db, subject, 0)

	} else if strings.HasPrefix(callbackData, "control_") {
		controlElement := strings.TrimPrefix(callbackData, "control_")
		userControlElement[chatID] = controlElement
		log.Printf("User %v choose control element: %v", chatID, controlElement)
		handleGetElementNumbers(bot, update, chatID, db, userSubject[chatID], controlElement, 0)

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

		deleteMsg := tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID)
		if _, err := bot.Send(deleteMsg); err != nil {
			if !isUnmarshalBoolError(err) {
				log.Printf("Failed to delete message with button to %v: %v", chatID, err)
			}
		}
		SendMaterial(bot, chatID, db, subject, controlElement, number)
	} else if strings.HasPrefix(callbackData, "subscriber_") {
		subscriberID := strings.TrimPrefix(callbackData, "subscriber_")
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("%v", subscriberID))
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending callback response: %v", err)
		}
	}

	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Callback error: %v", err)
	}
}

func isUnmarshalBoolError(err error) bool {
	return strings.Contains(err.Error(), "json: cannot unmarshal bool into Go value of type tgbotapi.Message")
}
