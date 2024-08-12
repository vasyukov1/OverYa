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
	isBroadcastMode   = false
	attachmentQueue   = []interface{}{}
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

			if chatID == adminMain && isBroadcastMode {
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
					isBroadcastMode = true
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

func handleAdminBroadcast(bot *tgbotapi.BotAPI, message *tgbotapi.Message, callback tgbotapi.Update, db *database.DB) {
	chatID := message.Chat.ID

	if message.Text != "" && broadcastMsg[chatID] == "" {
		broadcastMsg[chatID] = message.Text
		msg := tgbotapi.NewMessage(chatID, "Attach media (photo, video, file) and send /ok when done.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	} else if message.Text == "/ok" {
		var isDescriptionKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Yes", "Yes"),
				tgbotapi.NewInlineKeyboardButtonData("No", "/send"),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "Do you need description?")
		msg.ReplyMarkup = isDescriptionKeyboard
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	} else if message.Text == "/send" {
		isBroadcastMode = false
		broadcast(bot, broadcastMsg[chatID], attachmentQueue, description[chatID], db)
		description[chatID] = ""
		broadcastMsg[chatID] = ""
		attachmentQueue = []interface{}{}
		msg := tgbotapi.NewMessage(chatID, "Broadcast sent to all subscribers.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	} else if callback.CallbackQuery != nil && callback.CallbackQuery.Data == "Yes" {
		callback := tgbotapi.NewCallback(callback.CallbackQuery.ID, callback.CallbackQuery.Data)
		if _, err := bot.Request(callback); err != nil {
			log.Printf("Callback error: %v", err)
		}
		isDescriptionMode[chatID] = true
		msg := tgbotapi.NewMessage(chatID, "Send the description")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
	} else if isDescriptionMode[chatID] {
		description[chatID] = message.Text
		isDescriptionMode[chatID] = false
		msg := tgbotapi.NewMessage(chatID, "Tap /send")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
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

func handleMediaAttachment(message *tgbotapi.Message, db *database.DB) {
	//var media []interface{}
	var fileID string

	if message.Photo != nil {
		fileID = message.Photo[len(message.Photo)-1].FileID
		photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
		attachmentQueue = append(attachmentQueue, photo)
	} else if message.Video != nil {
		fileID = message.Video.FileID
		video := tgbotapi.NewInputMediaVideo(tgbotapi.FileID(fileID))
		attachmentQueue = append(attachmentQueue, video)
	} else if message.Document != nil {
		fileID = message.Document.FileID
		document := tgbotapi.NewInputMediaDocument(tgbotapi.FileID(fileID))
		attachmentQueue = append(attachmentQueue, document)
	}
	//if fileID != "" {
	//	if err := db.AddMaterial(broadcastMsg, fileID); err == nil {
	//		attachmentQueue = append(attachmentQueue, media)
	//	} else {
	//		log.Printf("Failed to add attachment to %v: %v\n", fileID, err)
	//	}
	//}
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
