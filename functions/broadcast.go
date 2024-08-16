package functions

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
	"strconv"
	"strings"
)

var (
	userSubject        = make(map[int64]string)
	userControlElement = make(map[int64]string)
	broadcastMsg       = make(map[int64]string)
	description        = make(map[int64]string)
	isDescriptionMode  = make(map[int64]bool)
	attachmentQueue    = make(map[int64][]interface{})
)

const MaxMediaGroupSize = 10

func HandleAdminBroadcast(bot *tgbotapi.BotAPI, message *tgbotapi.Message, update tgbotapi.Update, db *database.DB, telegramChannel int64, isBroadcastMode *map[int64]bool) {
	chatID := message.Chat.ID
	if update.CallbackQuery != nil {
		HandleCallbackQuery(bot, update, db, telegramChannel, isBroadcastMode)
		return
	}
	if message.Text != "" && broadcastMsg[chatID] == "" {
		broadcastMsg[chatID] = message.Text
		promptForAttachments(bot, chatID)
		return
	}
	if message.Text == "/ok" {
		promptForDescriptionChoice(bot, chatID)
		return
	}
	if isDescriptionMode[chatID] {
		description[chatID] = message.Text
		sendBroadcast(bot, chatID, db, telegramChannel, isBroadcastMode)
		isDescriptionMode[chatID] = false
		return
	}
	handleMediaAttachments(chatID, message)
}

func SendMaterial(bot *tgbotapi.BotAPI, chatID int64, db *database.DB, subject, controlElement string, number int) {
	files, description, err := db.GetMaterial(subject, controlElement, number)
	if err != nil {
		log.Printf("Error getting material: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Failed to retrieve the material.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	if len(files) == 0 {
		msg := tgbotapi.NewMessage(chatID, "There are not materials.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("There are not materials for %v: %v", chatID, err)
		}
		return
	}

	var mediaGroup []interface{}
	var hasDocumentsOrVideos bool

	for i, fileID := range files {
		if strings.HasSuffix(fileID, ".jpg") || strings.HasSuffix(fileID, ".jpeg") || strings.HasSuffix(fileID, ".png") {
			media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
			mediaGroup = append(mediaGroup, media)
			if i == 0 && description != "" {
				media.Caption = description
			}
		} else {
			media := tgbotapi.NewInputMediaDocument(tgbotapi.FileID(fileID))
			mediaGroup = append(mediaGroup, media)
			hasDocumentsOrVideos = true
		}

	}

	if hasDocumentsOrVideos {
		if len(mediaGroup) > 0 {
			group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
			if _, err := bot.Send(group); err != nil {
				log.Printf("Failed to send media group to %v: %v\n", chatID, err)
			}
		}
	} else {
		if len(mediaGroup) > 0 {
			group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
			if _, err := bot.Send(group); err != nil {
				log.Printf("Failed to send photo group to %v: %v\n", chatID, err)
			}
		}
	}
	if description != "" {
		msg := tgbotapi.NewMessage(chatID, description)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send description message to %v: %v\n", chatID, err)
		}
	}

}

func sendMediaGroup(bot *tgbotapi.BotAPI, chatID int64, mediaGroup []interface{}, hasDocumentsOrVideos bool) {
	if hasDocumentsOrVideos {
		if len(mediaGroup) > 0 {
			group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
			if _, err := bot.Send(group); err != nil {
				log.Printf("Failed to send media group to %v: %v\n", chatID, err)
			}
		}
	} else {
		if len(mediaGroup) > 0 {
			group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
			if _, err := bot.Send(group); err != nil {
				log.Printf("Failed to send photo group to %v: %v\n", chatID, err)
			}
		}
	}
}

func SendMaterialSearch(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	files, description, err := db.GetMaterialSearch(chatID)
	if err != nil {
		log.Printf("Error getting material: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Failed to retrieve the material.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	if len(files) == 0 {
		msg := tgbotapi.NewMessage(chatID, "There are not materials.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("There are not materials for %v: %v", chatID, err)
		}
		return
	}

	var mediaGroup []interface{}
	var hasDocumentsOrVideos bool

	for i, fileID := range files {
		//media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
		//mediaGroup = append(mediaGroup, media)
		if strings.HasSuffix(fileID, ".jpg") || strings.HasSuffix(fileID, ".jpeg") || strings.HasSuffix(fileID, ".png") {
			media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
			mediaGroup = append(mediaGroup, media)
			if i == 0 && description != "" {
				media.Caption = description
			}
		} else {
			media := tgbotapi.NewInputMediaDocument(tgbotapi.FileID(fileID))
			mediaGroup = append(mediaGroup, media)
			if i == 0 && description != "" {
				media.Caption = description
			}
			hasDocumentsOrVideos = true
		}

	}

	if hasDocumentsOrVideos {
		if len(mediaGroup) > 0 {
			group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
			if _, err := bot.Send(group); err != nil {
				log.Printf("Failed to send media group to %v: %v\n", chatID, err)
			}
		}
	} else {
		if len(mediaGroup) > 0 {
			group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
			if _, err := bot.Send(group); err != nil {
				log.Printf("Failed to send photo group to %v: %v\n", chatID, err)
			}
		}
	}

	//for _, fileID := range files {
	//	media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
	//	mediaGroup = append(mediaGroup, media)
	//}
	//
	//if len(mediaGroup) > 0 {
	//	group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
	//	if _, err := bot.Send(group); err != nil {
	//		log.Printf("Failed to send media group to %v: %v\n", chatID, err)
	//	}
	//}
	//
	//if description != "" {
	//	msg := tgbotapi.NewMessage(chatID, description)
	//	if _, err := bot.Send(msg); err != nil {
	//		log.Printf("Failed to send message to %v: %v\n", chatID, err)
	//	}
	//}
}

func broadcast(bot *tgbotapi.BotAPI, message string, attachments []interface{}, description string, db *database.DB, telegramChannel int64) {
	subscribers := db.GetSubscribers()
	for chatID := range subscribers {
		if len(attachments) > 0 {
			mediaGroup := tgbotapi.NewMediaGroup(chatID, attachments)
			sentMessages, err := bot.SendMediaGroup(mediaGroup)
			if err != nil {
				log.Printf("Send media group error to %v: %v\n", chatID, err)
			} else {
				for _, sentMsg := range sentMessages {
					log.Printf("Sent message ID: %v to chat ID: %v", sentMsg.MessageID, chatID)
				}
			}
		}

		subjectWithDescription := message + "\n\n" + description
		msg := tgbotapi.NewMessage(chatID, subjectWithDescription)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v\n", chatID, err)
		}
	}

	var links []string
	for _, attachment := range attachments {
		if attachment == nil {
			continue
		}
		var sentMsg tgbotapi.Message
		var err error

		switch media := attachment.(type) {
		case tgbotapi.InputMediaPhoto:
			photoMsg := tgbotapi.NewPhoto(telegramChannel, media.Media)
			sentMsg, err = bot.Send(photoMsg)
		case tgbotapi.InputMediaVideo:
			videoMsg := tgbotapi.NewVideo(telegramChannel, media.Media)
			sentMsg, err = bot.Send(videoMsg)
		case tgbotapi.InputMediaDocument:
			docMsg := tgbotapi.NewDocument(telegramChannel, media.Media)
			sentMsg, err = bot.Send(docMsg)
		default:
			log.Printf("Unknown media type: %T", media)
			continue
		}

		if err != nil {
			log.Printf("Failed to send media to channel: %v", err)
			continue
		}
		if sentMsg.Document != nil {
			links = append(links, sentMsg.Document.FileID)
		} else if sentMsg.Photo != nil {
			links = append(links, sentMsg.Photo[0].FileID)
		}
	}
	parts := strings.Split(message, " ")
	if len(parts) == 3 {
		subject := strings.TrimSpace(parts[0])
		controlElement := strings.TrimSpace(parts[1])
		number, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			log.Printf("Error converting number to int: %v\n", err)
		}

		exists, err := db.SubjectExists(subject)
		if err != nil {
			log.Printf("Error checking subject existence: %v\n", err)
			return
		}

		if !exists {
			err = db.AddSubject(subject)
			if err != nil {
				log.Printf("Error adding subject: %v\n", err)
				return
			}
			log.Printf("Added subject: %v\n", subject)
		} else {
			log.Printf("Subject exists: %v\n", subject)
		}

		err = db.AddMaterial(subject, controlElement, number, links, description)
		if err != nil {
			return
		}

	} else {
		log.Printf("Invalid message format: %v", len(parts))
	}
}

func promptForDescriptionChoice(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Do you want to add a description?")
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", "yes_description"),
			tgbotapi.NewInlineKeyboardButtonData("No", "no_description"),
		),
	)
	msg.ReplyMarkup = buttons
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

func HandleCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *database.DB, telegramChannel int64, isBroadcastMode *map[int64]bool) { //  callbackQuery *tgbotapi.CallbackQuery,
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackData := update.CallbackQuery.Data
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

	if strings.HasPrefix(callbackData, "subject_") {
		subject := strings.TrimPrefix(callbackData, "subject_")
		userSubject[chatID] = subject
		log.Printf("User %v choose subject: %v", chatID, subject)

		controlElements := db.GetControlElements(subject)
		if len(controlElements) == 0 {
			msg := tgbotapi.NewMessage(chatID, "No control elements found.")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Send message error to %v: %v", chatID, err)
			}
			return
		}

		var buttons [][]tgbotapi.InlineKeyboardButton
		for _, controlElement := range controlElements {
			button := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s", controlElement), fmt.Sprintf("control_%s", controlElement))
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
		}
		backButton := tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_subjects")
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(backButton))
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a control element:")
		editMsg.ReplyMarkup = &keyboard
		if _, err := bot.Send(editMsg); err != nil {
			log.Printf("Edit message error to %v: %v", chatID, err)
		}

	} else if strings.HasPrefix(callbackData, "control_") {
		controlElement := strings.TrimPrefix(callbackData, "control_")
		userControlElement[chatID] = controlElement
		log.Printf("User %v choose control element: %v", chatID, controlElement)

		elementNumbers := db.GetElementNumber(userSubject[chatID], controlElement)
		if len(elementNumbers) == 0 {
			msg := tgbotapi.NewMessage(chatID, "No element numbers found.")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Send message error to %v: %v", chatID, err)
			}
			return
		}

		var buttons [][]tgbotapi.InlineKeyboardButton
		for _, number := range elementNumbers {
			button := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%d", number), fmt.Sprintf("number_%d", number))
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
		}
		backButton := tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_controls")
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(backButton))
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a number:")
		editMsg.ReplyMarkup = &keyboard
		if _, err := bot.Send(editMsg); err != nil {
			log.Printf("Edit message error to %v: %v", chatID, err)
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
		subjects := db.GetSubjects()

		var buttons [][]tgbotapi.InlineKeyboardButton
		for _, subject := range subjects {
			button := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s", subject), fmt.Sprintf("subject_%s", subject))
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a subject:")
		editMsg.ReplyMarkup = &keyboard
		if _, err := bot.Send(editMsg); err != nil {
			return
		}

	} else if callbackData == "back_to_controls" {
		subject := userSubject[chatID]
		controlElements := db.GetControlElements(subject)
		if len(controlElements) == 0 {
			msg := tgbotapi.NewMessage(chatID, "No control elements found.")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Send message error to %v: %v", chatID, err)
			}
			return
		}

		var buttons [][]tgbotapi.InlineKeyboardButton
		for _, controlElement := range controlElements {
			button := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s", controlElement), fmt.Sprintf("control_%s", controlElement))
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
		}
		backButton := tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_subjects")
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(backButton))
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a control element:")
		editMsg.ReplyMarkup = &keyboard
		if _, err := bot.Send(editMsg); err != nil {
			log.Printf("Edit message error to %v: %v", chatID, err)
		}
	}

	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Callback error: %v", err)
	}
}

func promptForAttachments(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Attach media (photo, video, file) and send /ok when done.")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

func sendBroadcast(bot *tgbotapi.BotAPI, chatID int64, db *database.DB, telegramChannel int64, isBroadcastMode *map[int64]bool) {
	(*isBroadcastMode)[chatID] = false
	broadcast(bot, broadcastMsg[chatID], attachmentQueue[chatID], description[chatID], db, telegramChannel)
	description[chatID] = ""
	broadcastMsg[chatID] = ""
	attachmentQueue[chatID] = []interface{}{}
	msg := tgbotapi.NewMessage(chatID, "Broadcast sent to all subscribers.")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

func handleMediaAttachments(chatID int64, message *tgbotapi.Message) {
	if message.Photo != nil && len(message.Photo) > 0 {
		photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(message.Photo[len(message.Photo)-1].FileID))
		attachmentQueue[chatID] = append(attachmentQueue[chatID], photo)
	} else if message.Video != nil {
		video := tgbotapi.NewInputMediaVideo(tgbotapi.FileID(message.Video.FileID))
		attachmentQueue[chatID] = append(attachmentQueue[chatID], video)
	} else if message.Document != nil {
		document := tgbotapi.NewInputMediaDocument(tgbotapi.FileID(message.Document.FileID))
		attachmentQueue[chatID] = append(attachmentQueue[chatID], document)
	}
}

func handleGetSubjects(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	subjects := db.GetSubjects()
	if len(subjects) == 0 {
		msg := tgbotapi.NewMessage(chatID, "No subjects found.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send message error to %v: %v", chatID, err)
		}
		return
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, subject := range subjects {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s", subject), fmt.Sprintf("subject_%s", subject))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, "Select a subject:")
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Send message error to %v: %v", chatID, err)
	}
}
