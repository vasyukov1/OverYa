package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/config"
	"github.com/vasyukov1/Overbot/database"
	"log"
	"strconv"
	"strings"
)

var (
	broadcastMsg      = make(map[int64]string)
	description       = make(map[int64]string)
	isBroadcastMode   = make(map[int64]bool)
	attachmentQueue   = make(map[int64][]interface{})
	materialStep      = make(map[int64]string)
	isDescriptionMode = make(map[int64]bool)
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
	defer func(db *database.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Error closing database: %v\n", err)
		}
	}(db)

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

			if chatID == adminMain && isBroadcastMode[chatID] {
				handleAdminBroadcast(bot, update.Message, update, db)
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
					isBroadcastMode[chatID] = true
				} else {
					msg.Text = "You are not an admin"
				}
			case "get_materials":
				materialStep[chatID] = "awaiting_subject"
				msg.Text = "Please enter the subject name"
			default:
				msg.Text = "I don't know that command"
			}

			if update.Message.Command() == "" {

				if materialStep[chatID] != "" {
					switch materialStep[chatID] {
					case "awaiting_subject":
						msg.Text = "Please enter the control element (e.g., lecture, seminar) and its number"
						db.SetTempSubject(chatID, update.Message.Text)
						materialStep[chatID] = "awaiting_control_element"
					case "awaiting_control_element":
						msg.Text = "Please enter the number of element"
						db.SetTempControlElement(chatID, update.Message.Text)
						materialStep[chatID] = "awaiting_control_element"
					case "awaiting_element_number":
						msg.Text = "Please enter the number of element"
						elementNumberForGet, err := strconv.Atoi(update.Message.Text)
						if err != nil {
							msg.Text = "This element does not exist"
						} else {
							db.SetTempElementNumber(chatID, elementNumberForGet)
							materialStep[chatID] = ""
							sendMaterial(bot, chatID, db)
						}
					}
				}

				//switch update.Message.Text {
				//case "open":
				//	msg.ReplyMarkup = numericInlineKeyboard
				//default:
				//	msg.Text = "I don't understand you(("
				//}
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
func handleAdminBroadcast(bot *tgbotapi.BotAPI, message *tgbotapi.Message, update tgbotapi.Update, db *database.DB) {
	chatID := message.Chat.ID

	// Обработка сообщения с текстом, когда broadcastMsg пуст
	if message.Text != "" && broadcastMsg[chatID] == "" {
		broadcastMsg[chatID] = message.Text
		promptForAttachments(bot, chatID)
		return
	}
	// Обработка команды "/ok"
	if message.Text == "/ok" {
		promptForDescription(bot, chatID)
		return
	}
	// Обработка команды "/send"
	if message.Text == "/send" {
		sendBroadcast(bot, chatID, db)
		return
	}
	// Обработка inline-кнопок для описания
	if update.CallbackQuery != nil {
		handleCallbackQuery(bot, update, chatID)
		return
	}
	// Обработка текстовых сообщений для описания, если включен режим описания
	if isDescriptionMode[chatID] {
		handleDescriptionInput(bot, message, chatID)
		return
	}
	// Обработка медиа-файлов
	handleMediaAttachments(chatID, message)
}

func broadcast(bot *tgbotapi.BotAPI, message string, attachments []interface{}, description string, db *database.DB) {
	subscribers := db.GetSubscribers()
	for chatID := range subscribers {
		// Send media group if it exists.
		if len(attachments) > 0 {
			mediaGroup := tgbotapi.NewMediaGroup(chatID, attachments)
			if _, err := bot.Send(mediaGroup); err != nil {
				log.Printf("Send media group error to %v: %v\n", chatID, err)
			}
		}

		// Send type of materials and description.
		subjectWithDescription := message + description
		msg := tgbotapi.NewMessage(chatID, subjectWithDescription)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v\n", chatID, err)
		}

		// Here will be sending medias to Tg Channel
		// and receiving links to these materials
		// !PLUG!
		var links []string
		for _, attachment := range attachments {
			switch a := attachment.(type) {
			case tgbotapi.InputMediaPhoto:
				//links = append(links, a.MediaFileID)
			case tgbotapi.InputMediaVideo:
				//links = append(links, a.MediaFileID)
			case tgbotapi.InputMediaDocument:
				//links = append(links, a.MediaFileID)
			default:
				log.Printf("Unknown media type: %T", a)
			}
		}

		parts := strings.Split(message, ",")
		if len(parts) >= 3 {
			number, err := strconv.Atoi(parts[2])
			if err != nil {
				log.Printf("Error converting number to int: %v\n", err)
			} else {
				err := db.AddMaterial(parts[0], parts[1], number, links, description)
				if err != nil {
					return
				}
			}
		} else {
			log.Printf("Invalid message format: %s", message)
		}
	}
}

// Функция для запроса медиа-файлов у администратора
func promptForAttachments(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Attach media (photo, video, file) and send /ok when done.")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

// Функция для запроса описания у администратора
func promptForDescription(bot *tgbotapi.BotAPI, chatID int64) {
	var descriptionKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", "Yes"),
			tgbotapi.NewInlineKeyboardButtonData("No", "/send"),
		),
	)
	msg := tgbotapi.NewMessage(chatID, "Do you need a description?")
	msg.ReplyMarkup = descriptionKeyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

// Функция для обработки текстового ввода описания
func handleDescriptionInput(bot *tgbotapi.BotAPI, message *tgbotapi.Message, chatID int64) {
	description[chatID] = message.Text
	isDescriptionMode[chatID] = false
	msg := tgbotapi.NewMessage(chatID, "Tap /send")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, chatID int64) {
	if update.CallbackQuery.Data == "Yes" {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
		if _, err := bot.Request(callback); err != nil {
			log.Printf("Callback error: %v", err)
		}
		isDescriptionMode[chatID] = true
		msg := tgbotapi.NewMessage(chatID, "Send the description")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	}
}

func sendBroadcast(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	isBroadcastMode[chatID] = false
	broadcast(bot, broadcastMsg[chatID], attachmentQueue[chatID], description[chatID], db)
	description[chatID] = ""
	broadcastMsg[chatID] = ""
	attachmentQueue[chatID] = []interface{}{}
	msg := tgbotapi.NewMessage(chatID, "Broadcast sent to all subscribers.")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}

// Функция для обработки медиа-файлов
func handleMediaAttachments(chatID int64, message *tgbotapi.Message) {
	if message.Photo != nil {
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

func sendMaterial(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	files, description, err := db.GetMaterial(chatID)
	if err != nil {
		log.Printf("Error getting material: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Failed to retrieve the material.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	var mediaGroup []interface{}
	for _, fileID := range files {
		media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
		mediaGroup = append(mediaGroup, media)
	}

	if len(mediaGroup) > 0 {
		group := tgbotapi.NewMediaGroup(chatID, mediaGroup)
		if _, err := bot.Send(group); err != nil {
			log.Printf("Failed to send media group to %v: %v\n", chatID, err)
		}
	}

	if description != "" {
		msg := tgbotapi.NewMessage(chatID, description)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v\n", chatID, err)
		}
	}
}
