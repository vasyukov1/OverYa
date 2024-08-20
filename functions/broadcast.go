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

func broadcast(bot *tgbotapi.BotAPI, message string, attachments []interface{}, description string, db *database.DB, telegramChannel int64) {
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

		msg := tgbotapi.NewMessage(chatID, "")
		msg.Text = fmt.Sprintf("%v\n\n%v", message, description)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
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
			links = append(links, "document:"+sentMsg.Document.FileID)
		} else if sentMsg.Photo != nil {
			links = append(links, "photo:"+sentMsg.Photo[0].FileID)
		} else if sentMsg.Video != nil {
			links = append(links, "video:"+sentMsg.Video.FileID)
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
