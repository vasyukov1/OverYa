package functions

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"io"
	"log"
	"net/http"
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

//func SendMaterial(bot *tgbotapi.BotAPI, chatID int64, db *database.DB, subject, controlElement string, number int) {
//	files, description, err := db.GetMaterial(subject, controlElement, number)
//	if err != nil {
//		log.Printf("Error getting material: %v", err)
//		msg := tgbotapi.NewMessage(chatID, "Failed to retrieve the material.")
//		if _, err := bot.Send(msg); err != nil {
//			log.Printf("Failed to send message to %v: %v", chatID, err)
//		}
//		return
//	}
//
//	if len(files) == 0 {
//		msg := tgbotapi.NewMessage(chatID, "There are not materials.")
//		if _, err := bot.Send(msg); err != nil {
//			log.Printf("There are not materials for %v: %v", chatID, err)
//		}
//		return
//	}
//
//	var mediaGroup []interface{}
//	var hasDocumentsOrVideos bool
//
//	for i, fileID := range files {
//		if strings.HasSuffix(fileID, ".jpg") || strings.HasSuffix(fileID, ".jpeg") || strings.HasSuffix(fileID, ".png") {
//			media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
//			mediaGroup = append(mediaGroup, media)
//			if i == 0 && description != "" {
//				media.Caption = description
//			}
//		} else {
//			media := tgbotapi.NewInputMediaDocument(tgbotapi.FileID(fileID))
//			mediaGroup = append(mediaGroup, media)
//			hasDocumentsOrVideos = true
//		}
//
//	}
//
//	if hasDocumentsOrVideos {
//		if len(mediaGroup) > 0 {
//			group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
//			if _, err := bot.Send(group); err != nil {
//				log.Printf("Failed to send media group to %v: %v\n", chatID, err)
//			}
//		}
//	} else {
//		if len(mediaGroup) > 0 {
//			group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
//			if _, err := bot.Send(group); err != nil {
//				log.Printf("Failed to send photo group to %v: %v\n", chatID, err)
//			}
//		}
//	}
//	if description != "" {
//		msg := tgbotapi.NewMessage(chatID, description)
//		if _, err := bot.Send(msg); err != nil {
//			log.Printf("Failed to send description message to %v: %v\n", chatID, err)
//		}
//	}
//
//}

//	func sendMediaGroup(bot *tgbotapi.BotAPI, chatID int64, mediaGroup []interface{}, hasDocumentsOrVideos bool) {
//		if hasDocumentsOrVideos {
//			if len(mediaGroup) > 0 {
//				group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
//				if _, err := bot.Send(group); err != nil {
//					log.Printf("Failed to send media group to %v: %v\n", chatID, err)
//				}
//			}
//		} else {
//			if len(mediaGroup) > 0 {
//				group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
//				if _, err := bot.Send(group); err != nil {
//					log.Printf("Failed to send photo group to %v: %v\n", chatID, err)
//				}
//			}
//		}
//	}
func sendMediaGroup(bot *tgbotapi.BotAPI, chatID int64, mediaGroup []interface{}, description string) {
	group := make([]interface{}, len(mediaGroup))
	copy(group, mediaGroup)
	if len(group) > 0 {
		if media, ok := group[0].(*tgbotapi.InputMediaDocument); ok {
			media.Caption = description
		} else if media, ok := group[0].(*tgbotapi.InputMediaPhoto); ok {
			media.Caption = description
		}

		sentMessages, err := bot.SendMediaGroup(tgbotapi.NewMediaGroup(chatID, group))
		if err != nil {
			log.Printf("Send media group error to %v: %v\n", chatID, err)
		} else {
			for _, sentMsg := range sentMessages {
				log.Printf("Sent message ID: %v to chat ID: %v", sentMsg.MessageID, chatID)
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
}

func broadcast(bot *tgbotapi.BotAPI, message string, attachments []interface{}, description string, db *database.DB, telegramChannel int64) {
	subscribers := db.GetSubscribers()
	// test
	var photoVideoGroup []interface{}
	var documentGroup []interface{}

	for _, attachment := range attachments {
		switch media := attachment.(type) {
		case tgbotapi.InputMediaPhoto, tgbotapi.InputMediaVideo:
			photoVideoGroup = append(photoVideoGroup, media)
		case tgbotapi.InputMediaDocument:
			documentGroup = append(documentGroup, media)
		}
	}
	// end test
	for chatID := range subscribers {
		//if len(attachments) > 0 {
		//	mediaGroup := tgbotapi.NewMediaGroup(chatID, attachments)
		//	sentMessages, err := bot.SendMediaGroup(mediaGroup)
		//	if err != nil {
		//		log.Printf("Send media group error to %v: %v\n", chatID, err)
		//	} else {
		//		for _, sentMsg := range sentMessages {
		//			log.Printf("Sent message ID: %v to chat ID: %v", sentMsg.MessageID, chatID)
		//		}
		//	}
		//}
		if len(documentGroup) > 0 {
			sendMediaGroup(bot, chatID, documentGroup, description)
		}
		if len(photoVideoGroup) > 0 {
			sendMediaGroup(bot, chatID, photoVideoGroup, description)
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

//func HandleCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *database.DB, telegramChannel int64, isBroadcastMode *map[int64]bool) { //  callbackQuery *tgbotapi.CallbackQuery,
//	chatID := update.CallbackQuery.Message.Chat.ID
//	callbackData := update.CallbackQuery.Data
//	switch callbackData {
//	case "yes_description":
//		msg := tgbotapi.NewMessage(chatID, "Please provide the description")
//		if _, err := bot.Send(msg); err != nil {
//			log.Printf("Failed to send message to %v: %v", chatID, err)
//		}
//		isDescriptionMode[chatID] = true
//	case "no_description":
//		sendBroadcast(bot, chatID, db, telegramChannel, isBroadcastMode)
//	}
//
//	if strings.HasPrefix(callbackData, "subjects_page_") {
//		pageStr := strings.TrimPrefix(callbackData, "subjects_page_")
//		page, err := strconv.Atoi(pageStr)
//
//		if err != nil {
//			log.Printf("Invalid page number: %v", err)
//			return
//		}
//
//		handleGetSubjects(bot, update, chatID, db, page)
//
//		//Отвечаем на callback_query, чтобы убрать индикатор ожидания в клиенте
//		answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
//		if _, err := bot.Request(answer); err != nil {
//			log.Printf("Error sending callback response: %v", err)
//		}
//	} else if strings.HasPrefix(callbackData, "subject_") {
//		subject := strings.TrimPrefix(callbackData, "subject_")
//		userSubject[chatID] = subject
//		log.Printf("User %v choose subject: %v", chatID, subject)
//
//		controlElements := db.GetControlElements(subject)
//		if len(controlElements) == 0 {
//			msg := tgbotapi.NewMessage(chatID, "No control elements found.")
//			if _, err := bot.Send(msg); err != nil {
//				log.Printf("Send message error to %v: %v", chatID, err)
//			}
//			return
//		}
//
//		var buttons [][]tgbotapi.InlineKeyboardButton
//		for _, controlElement := range controlElements {
//			button := tgbotapi.NewInlineKeyboardButtonData(
//				fmt.Sprintf("%s", controlElement), fmt.Sprintf("control_%s", controlElement))
//			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
//		}
//		backButton := tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_subjects")
//		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(backButton))
//		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
//
//		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a control element:")
//		editMsg.ReplyMarkup = &keyboard
//		if _, err := bot.Send(editMsg); err != nil {
//			log.Printf("Edit message error to %v: %v", chatID, err)
//		}
//
//	} else if strings.HasPrefix(callbackData, "control_") {
//		controlElement := strings.TrimPrefix(callbackData, "control_")
//		userControlElement[chatID] = controlElement
//		log.Printf("User %v choose control element: %v", chatID, controlElement)
//
//		elementNumbers := db.GetElementNumber(userSubject[chatID], controlElement)
//		if len(elementNumbers) == 0 {
//			msg := tgbotapi.NewMessage(chatID, "No element numbers found.")
//			if _, err := bot.Send(msg); err != nil {
//				log.Printf("Send message error to %v: %v", chatID, err)
//			}
//			return
//		}
//
//		var buttons [][]tgbotapi.InlineKeyboardButton
//		for _, number := range elementNumbers {
//			button := tgbotapi.NewInlineKeyboardButtonData(
//				fmt.Sprintf("%d", number), fmt.Sprintf("number_%d", number))
//			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
//		}
//		backButton := tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_controls")
//		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(backButton))
//		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
//
//		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a number:")
//		editMsg.ReplyMarkup = &keyboard
//		if _, err := bot.Send(editMsg); err != nil {
//			log.Printf("Edit message error to %v: %v", chatID, err)
//		}
//
//	} else if strings.HasPrefix(callbackData, "number_") {
//		numberStr := strings.TrimPrefix(callbackData, "number_")
//		number, err := strconv.Atoi(numberStr)
//		if err != nil {
//			log.Println("Invalid number:", numberStr)
//			return
//		}
//		log.Printf("User %v choose element number: %v", chatID, number)
//
//		subject := userSubject[chatID]
//		controlElement := userControlElement[chatID]
//		SendMaterial(bot, chatID, db, subject, controlElement, number)
//
//	} else if callbackData == "back_to_subjects" {
//		subjects := db.GetSubjects()
//
//		var buttons [][]tgbotapi.InlineKeyboardButton
//		for _, subject := range subjects {
//			button := tgbotapi.NewInlineKeyboardButtonData(
//				fmt.Sprintf("%s", subject), fmt.Sprintf("subject_%s", subject))
//			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
//		}
//		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
//
//		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a subject:")
//		editMsg.ReplyMarkup = &keyboard
//
//		if _, err := bot.Send(editMsg); err != nil {
//			return
//		}
//
//	} else if callbackData == "back_to_controls" {
//		subject := userSubject[chatID]
//		controlElements := db.GetControlElements(subject)
//		if len(controlElements) == 0 {
//			msg := tgbotapi.NewMessage(chatID, "No control elements found.")
//			if _, err := bot.Send(msg); err != nil {
//				log.Printf("Send message error to %v: %v", chatID, err)
//			}
//			return
//		}
//
//		var buttons [][]tgbotapi.InlineKeyboardButton
//		for _, controlElement := range controlElements {
//			button := tgbotapi.NewInlineKeyboardButtonData(
//				fmt.Sprintf("%s", controlElement), fmt.Sprintf("control_%s", controlElement))
//			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
//		}
//		backButton := tgbotapi.NewInlineKeyboardButtonData("Back", "back_to_subjects")
//		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(backButton))
//		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
//
//		editMsg := tgbotapi.NewEditMessageText(chatID, update.CallbackQuery.Message.MessageID, "Select a control element:")
//		editMsg.ReplyMarkup = &keyboard
//		if _, err := bot.Send(editMsg); err != nil {
//			log.Printf("Edit message error to %v: %v", chatID, err)
//		}
//	}
//
//	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
//	if _, err := bot.Request(callback); err != nil {
//		log.Printf("Callback error: %v", err)
//	}
//}

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

func SendMaterial(bot *tgbotapi.BotAPI, chatID int64, db *database.DB, subject, controlElement string, number int) {
	// Получаем материалы и описание из базы данных
	files, description, err := db.GetMaterial(subject, controlElement, number)
	if err != nil {
		log.Printf("Error getting materials: %v", err)
		return
	}

	// Если материалов нет, отправляем сообщение "нет материалов"
	if len(files) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Нет материалов")
		bot.Send(msg)
		return
	}

	// Разделение файлов на фото/видео и документы
	var photoVideoFiles []string
	var documentFiles []string

	for _, fileID := range files {
		if isPhoto(bot, fileID) {
			photoVideoFiles = append(photoVideoFiles, fileID)
		} else {
			documentFiles = append(documentFiles, fileID)
		}
	}

	// Функция для отправки медиа-групп
	sendMediaGroup1 := func(files []string, description string, isDocument bool) {
		mediaGroup := make([]interface{}, len(files))

		for i, fileID := range files {
			if isDocument {
				media := tgbotapi.NewInputMediaDocument(tgbotapi.FileID(fileID))
				if i == 0 {
					media.Caption = description
				}
				mediaGroup[i] = media
			} else {
				media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
				if i == 0 {
					media.Caption = description
				}
				mediaGroup[i] = media
			}
		}

		if _, err := bot.Send(tgbotapi.NewMediaGroup(chatID, mediaGroup)); err != nil {
			log.Printf("Failed to send media group to %v: %v", chatID, err)
		}
	}

	// Отправка фото и видео группами по 10
	for i := 0; i < len(photoVideoFiles); i += 10 {
		end := i + 10
		if end > len(photoVideoFiles) {
			end = len(photoVideoFiles)
		}
		sendMediaGroup1(photoVideoFiles[i:end], description, false)
	}

	// Отправка документов группами по 10
	for i := 0; i < len(documentFiles); i += 10 {
		end := i + 10
		if end > len(documentFiles) {
			end = len(documentFiles)
		}
		sendMediaGroup1(documentFiles[i:end], description, true)
	}

	//// Получаем материалы и описание из базы данных
	//files, description, err := db.GetMaterial(subject, controlElement, number)
	//if err != nil {
	//	log.Printf("Error getting materials: %v", err)
	//	return
	//}
	//log.Printf("YAROSLAVA")
	//
	//// Если материалов нет, отправляем сообщение "нет материалов"
	//if len(files) == 0 {
	//	msg := tgbotapi.NewMessage(chatID, "Нет материалов")
	//	bot.Send(msg)
	//	return
	//}
	//
	//log.Printf("YAROSLAVA 2")
	//// Функция для создания группы материалов
	//createMediaGroup := func(files []string, description string) ([]interface{}, bool) {
	//	mediaGroup := []interface{}{}
	//	hasDocument := false
	//	log.Printf("YAROSLAVA CREATE")
	//
	//	for i, fileID := range files {
	//		if isPhoto(bot, fileID) {
	//			media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
	//			if i == 0 {
	//				media.Caption = description
	//			}
	//			mediaGroup = append(mediaGroup, media)
	//		} else {
	//			media := tgbotapi.NewInputMediaDocument(tgbotapi.FileID(fileID))
	//			if i == 0 {
	//				media.Caption = description
	//			}
	//			mediaGroup = append(mediaGroup, media)
	//			hasDocument = true
	//		}
	//	}
	//	log.Printf("YAROSLAVA 3")
	//	// Если есть хотя бы один документ, все материалы отправляются как документы
	//	//if hasDocument {
	//	//	log.Printf("YAROSLAVA HAS DOCUMENT")
	//	//	for i := range mediaGroup {
	//	//		if _, ok := mediaGroup[i].(*tgbotapi.InputMediaPhoto); ok {
	//	//			mediaGroup[i] = tgbotapi.NewInputMediaDocument(tgbotapi.FileID(files[i]))
	//	//		}
	//	//	}
	//	//}
	//
	//	return mediaGroup, hasDocument
	//}
	//
	//log.Printf("YAROSLAVA 4")
	//// Отправка материалов группами по 10 элементов
	//for i := 0; i < len(files); i += 10 {
	//	end := i + 10
	//	if end > len(files) {
	//		end = len(files)
	//	}
	//	mediaGroup, hasDocument := createMediaGroup(files[i:end], description)
	//	if hasDocument {
	//		for j := range mediaGroup {
	//			if _, ok := mediaGroup[j].(*tgbotapi.InputMediaPhoto); ok {
	//				mediaGroup[j] = tgbotapi.NewInputMediaDocument(tgbotapi.FileID(files[i+j]))
	//			}
	//		}
	//	}
	//	if _, err := bot.Send(tgbotapi.NewMediaGroup(chatID, mediaGroup)); err != nil {
	//		log.Printf("Failed to send material message to %v: %v", chatID, err)
	//	}
	//}
}

func isPhoto(bot *tgbotapi.BotAPI, fileID string) bool {
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		return false
	}

	// Получаем прямой URL файла
	fileURL := file.Link(bot.Token)

	// Делаем HTTP-запрос для получения заголовков файла
	resp, err := http.Head(fileURL)
	if err != nil {
		log.Printf("Error making HEAD request: %v", err)
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing body: %v", err)
		}
	}(resp.Body)

	// Проверяем Content-Type в заголовках
	contentType := resp.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "image/")
}
