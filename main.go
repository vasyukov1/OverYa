package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/config"
	"github.com/vasyukov1/Overbot/database"
	"log"
)

var (
	broadcastMsg    = ""
	isBroadcastMode = false
	attachmentQueue = []interface{}{}
)

var numericInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Overmindv", "https://t.me/overmindv"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Algebra", "Algebra"),
		tgbotapi.NewInlineKeyboardButtonURL("Les", "https://t.me/forest"),
		tgbotapi.NewInlineKeyboardButtonData("Calculus", "Calculus"),
	),
)

func main() {
	cfg := config.LoadConfig()

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Panic(err)
	}

	// We need to settle it: = false
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	adminMain := cfg.AdminID

	db, err := database.NewDB()
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}
	defer db.Close()

	err = db.CreateTables()
	if err != nil {
		log.Fatalf("Error creating tables: %v\n", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			msg := tgbotapi.NewMessage(chatID, "")

			if !db.IsSubscriber(chatID) {
				db.AddSubscriber(chatID)
			}

			if chatID == adminMain && isBroadcastMode {
				handleAdminBroadcast(bot, update.Message, db)
				continue
			}

			switch update.Message.Command() {
			case "start":
				msg.Text = "Hello, HSE Student!"
			case "help":
				msg.Text = "Usage: /start, /help, /broadcast"
			case "broadcast":
				if chatID == adminMain {
					msg.Text = "Please enter the subject and control element, e.g., 'Algebra lecture 2'."
					isBroadcastMode = true
				} else {
					msg.Text = "You are not an admin"
				}
			default:
				msg.Text = "I don't know that command"
			}

			if update.Message.Command() == "" {
				switch update.Message.Text {
				case "open":
					msg.ReplyMarkup = numericInlineKeyboard
				default:
					msg.Text = "I don't understand you(("
				}
			}

			if _, err := bot.Send(msg); err != nil {
				log.Printf("Send message error to %v: %v", chatID, err)
			}

		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				log.Panic(err)
			}

			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		}
	}
}

func handleAdminBroadcast(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *database.DB) {
	chatID := message.Chat.ID

	if message.Text != "" && broadcastMsg == "" {
		broadcastMsg = message.Text
		msg := tgbotapi.NewMessage(chatID, "Attach media (photo, video, file) and send /ok when done.")
		bot.Send(msg)
	} else if message.Text == "/ok" {
		isBroadcastMode = false
		broadcast(bot, broadcastMsg, attachmentQueue, db)
		broadcastMsg = ""
		attachmentQueue = []interface{}{}
		msg := tgbotapi.NewMessage(chatID, "Broadcast sent to all subscribers.")
		bot.Send(msg)
	} else if message.Photo != nil || message.Video != nil || message.Document != nil {
		if message.Photo != nil {
			photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(message.Photo[len(message.Photo)-1].FileID))
			attachmentQueue = append(attachmentQueue, photo)
		} else if message.Video != nil {
			video := tgbotapi.NewInputMediaVideo(tgbotapi.FileID(message.Video.FileID))
			attachmentQueue = append(attachmentQueue, video)
		} else if message.Document != nil {
			document := tgbotapi.NewInputMediaDocument(tgbotapi.FileID(message.Document.FileID))
			attachmentQueue = append(attachmentQueue, document)
		}
	}
}

func broadcast(bot *tgbotapi.BotAPI, message string, attachments []interface{}, db *database.DB) {
	subscribers := db.GetSubscribers()
	for chatID := range subscribers {
		msg := tgbotapi.NewMessage(chatID, message)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v\n", chatID, err)
		}

		if len(attachments) > 0 {
			mediaGroup := tgbotapi.NewMediaGroup(chatID, attachments)
			if _, err := bot.Send(mediaGroup); err != nil {
				log.Printf("Send media group error to %v: %v\n", chatID, err)
			}
		}
	}
}
