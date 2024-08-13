package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/config"
	"github.com/vasyukov1/Overbot/database"
	"github.com/vasyukov1/Overbot/functions"
	"github.com/vasyukov1/Overbot/users/subscribers"
	"log"
	"strconv"
	"strings"
)

var (
	isBroadcastMode        = make(map[int64]bool)
	materialStep           = make(map[int64]string)
	inProcess              = make(map[int64]bool)
	isDeleteSubscriberMode = false
)

func main() {
	cfg := config.LoadConfig()

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Panic(err)
	}
	// We need to settle it: = false
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	adminMain := cfg.AdminID
	telegramChannel := cfg.TelegramChannel

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

			//if !db.IsSubscriber(chatID) {
			//	db.AddSubscriber(chatID)
			//}

			if !db.IsSubscriber(chatID) {
				if !inProcess[chatID] {
					msg.Text = "If you want to become a subscriber, click on /become_subscriber"
					switch update.Message.Command() {
					case "become_subscriber":
						msg.Text = "Your request is processed"
						inProcess[chatID] = true
						users.SendSubscribeRequest(bot, chatID, adminMain, db)
					}
				} else {
					msg.Text = "Your request is processed"
				}

			} else {

				if chatID == adminMain && isBroadcastMode[chatID] {
					functions.HandleAdminBroadcast(bot, update.Message, update, db, telegramChannel, &isBroadcastMode)
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
				case "delete_subscriber":
					if chatID == adminMain {
						msg.Text = "Send subscriber's ID"
						isDeleteSubscriberMode = true
					}
				case "requests":

				}

				if update.Message.Command() == "" {

					if materialStep[chatID] != "" {
						switch materialStep[chatID] {
						case "awaiting_subject":
							msg.Text = "Please enter the control element (e.g., lecture, seminar)"
							db.SetTempSubject(chatID, update.Message.Text)
							materialStep[chatID] = "awaiting_control_element"
						case "awaiting_control_element":
							msg.Text = "Please enter the number of element"
							db.SetTempControlElement(chatID, update.Message.Text)
							materialStep[chatID] = "awaiting_element_number"
						case "awaiting_element_number":
							elementNumberForGet, err := strconv.Atoi(update.Message.Text)
							if err != nil {
								msg.Text = "This element does not exist"
							} else {
								db.SetTempElementNumber(chatID, elementNumberForGet)
								materialStep[chatID] = ""
								functions.SendMaterial(bot, chatID, db)
							}
						}
					}
					if isDeleteSubscriberMode && chatID == adminMain {
						deleteID, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
						if err != nil {
							log.Printf("Error converting delete_subscriber_id to int: %v\n", err)
							msg.Text = "We can't delete this subscriber"
						} else {
							users.DeleteSubscriber(bot, int64(deleteID), adminMain, db)
							msg.Text = "Deleting..."
						}

					}
				}
			}

			if _, err := bot.Send(msg); err != nil {
				log.Printf("Send message error to %v: %v", chatID, err)
			}

		} else if update.CallbackQuery != nil && db.IsSubscriber(update.CallbackQuery.From.ID) {
			functions.HandleCallbackQuery(bot, update.CallbackQuery, db, telegramChannel, &isBroadcastMode)
		}
	}
}

func handleSubscriberRequests(bot *tgbotapi.BotAPI, chatID int64, db *database.DB) {
	if !db.IsAdmin(chatID) {
		msg := tgbotapi.NewMessage(chatID, "У вас нет прав администратора.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	requests, err := db.GetSubscriberRequests()
	if err != nil {
		log.Printf("Failed to get subscriber requests: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Не удалось получить заявки на подписку.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	if len(requests) == 0 {
		msg := tgbotapi.NewMessage(chatID, "There are no requests.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v", chatID, err)
		}
		return
	}

	buttons := make([]tgbotapi.InlineKeyboardButton, 0, len(requests))
	for _, request := range requests {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf(strconv.Itoa(int(request))), fmt.Sprintf("request_%d", request))
		buttons = append(buttons, button)
	}

	msg := tgbotapi.NewMessage(chatID, "Subscriber Requests:")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message to %v: %v", chatID, err)
	}
}
