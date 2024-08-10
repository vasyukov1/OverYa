package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/config"
	"log"
)

// Temporary storage for subscribers
var (
	subscribers     = make(map[int64]bool)
	broadcastMsg    = ""
	isBroadcastMode = false
	attachmentQueue = []tgbotapi.Chattable{}
)

const subscribersFile = "subscribers.txt"

const adminMain = 2088252813 //cfg.MainAdmin

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

	bot.Debug = true // We need to settle it: = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	//loadSubscribers()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			msg := tgbotapi.NewMessage(chatID, "")

			// Here will be a check of subscription
			if _, exist := subscribers[update.Message.Chat.ID]; !exist {
				subscribers[chatID] = true
				//saveSubscribers()
			}

			if chatID == adminMain && isBroadcastMode {
				//handleAdminBroadcast(bot, update.Message)
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

			switch update.Message.Text {
			case "open":
				msg.ReplyMarkup = numericInlineKeyboard
			default:
				msg.Text = "I don't understand you(("
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

func broadcast(bot *tgbotapi.BotAPI, message string) {
	for chatID := range subscribers {
		msg := tgbotapi.NewMessage(chatID, message)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message to %v: %v\n", chatID, err)
		}
	}

}
