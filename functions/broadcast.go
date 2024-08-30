package functions

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
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

func HandleAdminBroadcast(bot *tgbotapi.BotAPI, message *tgbotapi.Message, update tgbotapi.Update, db *database.DB, telegramChannel int64, isBroadcastMode *map[int64]bool) {
	chatID := message.Chat.ID

	if update.CallbackQuery != nil {
		HandleCallbackQuery(bot, update, db, telegramChannel, isBroadcastMode)
		return
	}

	if message.Text != "" && broadcastMsg[chatID] == "" {
		broadcastMsg[chatID] = message.Text
		next := true

		parts := strings.Split(broadcastMsg[chatID], " ")
		if len(parts) == 3 {
			subject := strings.TrimSpace(parts[0])
			controlElement := strings.TrimSpace(parts[1])
			number := strings.TrimSpace(parts[2])

			exists, err := db.SubjectExists(subject)
			if err != nil {
				log.Printf("Error checking subject existence: %v\n", err)
			}

			if !exists {
				err = db.AddSubject(subject)
				if err != nil {
					log.Printf("Error adding subject: %v\n", err)
					next = false
				}
				log.Printf("Added subject: %v\n", subject)
			} else {
				log.Printf("Subject exists: %v\n", subject)
			}

			if db.IsMaterialExists(subject, controlElement, number) {
				//buttons := []tgbotapi.InlineKeyboardButton{
				//	tgbotapi.NewInlineKeyboardButtonData("Да", fmt.Sprintf("edit_material_%v_%v_%v", subject, controlElement, number)),
				//	tgbotapi.NewInlineKeyboardButtonData("Нет", "do_not_edit_material"),
				//}
				//keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
				//editMsg := tgbotapi.NewMessage(chatID, "Хотите редактировать его?")
				//editMsg.ReplyMarkup = &keyboard
				//if _, err := bot.Send(editMsg); err != nil {
				//	log.Printf("Failed to send message with buttons to %v: %v", chatID, err)
				//}
				msg := tgbotapi.NewMessage(chatID, "")
				msg.Text = fmt.Sprintf("Material `%v %v %v` is already existing", subject, controlElement, number)
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Failed to send message to %v: %v", chatID, err)
				}

				broadcastMsg[chatID] = ""
				(*isBroadcastMode)[chatID] = false
				attachmentQueue[chatID] = nil

				return
			}
			//GoToMain(chatID, db, bot)
		} else {
			msg := tgbotapi.NewMessage(chatID, "")
			msg.Text = fmt.Sprintf("Неверное название материала")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Failed to send message to %v: %v", chatID, err)
			}
			//GoToMain(chatID, db, bot)
		}
		if next {
			promptForAttachments(bot, chatID)
		}
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

func broadcast(bot *tgbotapi.BotAPI, adminID int64, message string, attachments []interface{}, description string, db *database.DB, telegramChannel int64) bool {
	subscribers := db.GetSubscribers()
	var groupPhoto []interface{}
	var groupDocument []interface{}
	var groupVideo []interface{}

	for _, attachment := range attachments {
		switch media := attachment.(type) {
		case tgbotapi.InputMediaPhoto:
			groupPhoto = append(groupPhoto, media)
		case tgbotapi.InputMediaDocument:
			groupDocument = append(groupDocument, media)
		case tgbotapi.InputMediaVideo:
			groupVideo = append(groupVideo, media)
		}
	}

	parts := strings.Split(message, " ")
	if len(parts) != 3 {
		return false
	}
	subject := strings.TrimSpace(parts[0])
	controlElement := strings.TrimSpace(parts[1])
	number := strings.TrimSpace(parts[2])
	links := saveInChannel(bot, telegramChannel, attachments)
	err := db.AddMaterial(subject, controlElement, number, links, description)
	if err != nil {
		return false
	}

	for chatID := range subscribers {
		sendMediaGroup := func(mediaGroup []interface{}) {
			if len(mediaGroup) > 0 {
				msg := tgbotapi.NewMediaGroup(chatID, mediaGroup)
				_, err := bot.Send(msg)
				if err != nil {
					if !isUnmarshalError(err) {
						log.Printf("Failed to send media group to %v: %v", chatID, err)
					}
				}
			}
		}
		msg := tgbotapi.NewMessage(chatID, "")
		msg.Text = fmt.Sprintf("РАССЫЛКА")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		for i := 0; i < len(groupPhoto); i += 10 {
			end := i + 10
			if end > len(groupPhoto) {
				end = len(groupPhoto)
			}
			sendMediaGroup(groupPhoto[i:end])
		}
		for i := 0; i < len(groupVideo); i += 10 {
			end := i + 10
			if end > len(groupVideo) {
				end = len(groupVideo)
			}
			sendMediaGroup(groupVideo[i:end])
		}
		for i := 0; i < len(groupDocument); i += 10 {
			end := i + 10
			if end > len(groupDocument) {
				end = len(groupDocument)
			}
			sendMediaGroup(groupDocument[i:end])
		}

		msg.Text = fmt.Sprintf("%v\n\n%v", message, description)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	}
	return true
}

func promptForDescriptionChoice(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Хотите добавить описание?")
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", "yes_description"),
			tgbotapi.NewInlineKeyboardButtonData("Нет", "no_description"),
		),
	)
	msg.ReplyMarkup = buttons
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

func promptForAttachments(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Прикрепите медиа (фото, видео, файлы) и после этого нажмите /ok")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

func sendBroadcast(bot *tgbotapi.BotAPI, chatID int64, db *database.DB, telegramChannel int64, isBroadcastMode *map[int64]bool) {
	(*isBroadcastMode)[chatID] = false
	if broadcast(bot, chatID, broadcastMsg[chatID], attachmentQueue[chatID], description[chatID], db, telegramChannel) {
		description[chatID] = ""
		broadcastMsg[chatID] = ""
		attachmentQueue[chatID] = []interface{}{}
		msg := tgbotapi.NewMessage(chatID, "Рассылка отправлена всем подписчикам")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, "Ошибка рассылки")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	}
}

func handleMediaAttachments(chatID int64, message *tgbotapi.Message) {
	if message.Photo != nil && len(message.Photo) > 0 {
		fileID := message.Photo[len(message.Photo)-1].FileID
		photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
		attachmentQueue[chatID] = append(attachmentQueue[chatID], photo)
	} else if message.Video != nil {
		fileID := message.Video.FileID
		video := tgbotapi.NewInputMediaVideo(tgbotapi.FileID(fileID))
		attachmentQueue[chatID] = append(attachmentQueue[chatID], video)
	} else if message.Document != nil {
		fileID := message.Document.FileID
		document := tgbotapi.NewInputMediaDocument(tgbotapi.FileID(fileID))
		attachmentQueue[chatID] = append(attachmentQueue[chatID], document)
	}
}

func saveInChannel(bot *tgbotapi.BotAPI, telegramChannel int64, attachments []interface{}) []string {
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
			links = append(links, "document:"+sentMsg.Document.FileID)
		} else if sentMsg.Photo != nil {
			links = append(links, "photo:"+sentMsg.Photo[0].FileID)
		} else if sentMsg.Video != nil {
			links = append(links, "video:"+sentMsg.Video.FileID)
		}
	}
	return links
}

func HandleEditMaterial(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64, db *database.DB, subject, controlElement string, number int) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Редактирование материала %v %v %v", subject, controlElement, number))
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}

	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Название", fmt.Sprintf("edit_name_%v_%v_%v", subject, controlElement, number)),
		tgbotapi.NewInlineKeyboardButtonData("Медиа", fmt.Sprintf("edit_media_%v_%v_%v", subject, controlElement, number)),
		tgbotapi.NewInlineKeyboardButtonData("Описание", fmt.Sprintf("edit_description_%v_%v_%v", subject, controlElement, number)),
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
	msg.Text = "Выберите, что хотите изменить?"
	msg.ReplyMarkup = &keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message with buttons to %v: %v", chatID, err)
	}
}
