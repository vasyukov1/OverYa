package functions

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/database"
	"log"
	"strings"
)

func SendMaterial(bot *tgbotapi.BotAPI, chatID int64, db *database.DB, subject, controlElement string, number int) {
	files, description, err := db.GetMaterial(subject, controlElement, number)
	if err != nil {
		log.Printf("Error getting materials: %v", err)
		return
	}

	if len(files) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Нет материалов")
		if _, err := bot.Send(msg); err != nil {
			return
		}
		return
	}

	var groupPhoto []interface{}
	var groupDocument []interface{}
	var groupVideo []interface{}

	for _, file := range files {
		if isPhoto(file) {
			link := strings.TrimPrefix(file, "photo:")
			groupPhoto = append(groupPhoto, tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(link)))
		} else if isDocument(file) {
			link := strings.TrimPrefix(file, "document:")
			groupDocument = append(groupDocument, tgbotapi.NewInputMediaDocument(tgbotapi.FileID(link)))
		} else if isVideo(file) {
			link := strings.TrimPrefix(file, "video:")
			groupVideo = append(groupVideo, tgbotapi.NewInputMediaVideo(tgbotapi.FileID(link)))
		} else {
			log.Printf("We don't know this file: %v", file)
		}
	}

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
	msg.Text = fmt.Sprintf("%v %v %v\n\n%v", subject, controlElement, number, description)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

func isPhoto(fileID string) bool {
	return hasPrefix(fileID, "photo:")
}

func isVideo(fileID string) bool {
	return hasPrefix(fileID, "video:")
}

func isDocument(fileID string) bool {
	return hasPrefix(fileID, "document:")
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func isUnmarshalError(err error) bool {
	return strings.Contains(err.Error(), "json: cannot unmarshal array into Go value of type tgbotapi.Message")
}
