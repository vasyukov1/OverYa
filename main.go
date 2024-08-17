package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vasyukov1/Overbot/config"
	"github.com/vasyukov1/Overbot/database"
	"github.com/vasyukov1/Overbot/functions"
	"github.com/vasyukov1/Overbot/users/admins"
	"github.com/vasyukov1/Overbot/users/subscribers"
	"log"
	"strconv"
	"strings"
)

var (
	userSubject            = make(map[int64]string)
	userControlElement     = make(map[int64]string)
	isBroadcastMode        = make(map[int64]bool)
	materialStep           = make(map[int64]string)
	inProcessSubReq        = make(map[int64]bool)
	inProcessAdminReq      = make(map[int64]bool)
	isDeleteSubscriberMode = false
	isDeleteAdminMode      = false
	isDeleteMaterialMode   = false
	isDeleteSubjectMode    = false
)

func main() {
	cfg := config.LoadConfig()

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

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

	telegramChannel := cfg.TelegramChannel
	adminMain := cfg.AdminID
	subscribers.AddSubscriber(bot, adminMain, db)
	admins.AddAdmin(bot, adminMain, db)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			msg := tgbotapi.NewMessage(chatID, "")

			if !db.IsSubscriber(chatID) {
				if !inProcessSubReq[chatID] {
					msg.Text = "If you want to become a subscriber, click on /become_subscriber"
					switch update.Message.Command() {
					case "become_subscriber":
						msg.Text = "Your request is processed"
						inProcessSubReq[chatID] = true
						subscribers.SendSubscribeRequest(bot, chatID, adminMain, db, update.Message.Chat)
					}
				} else {
					msg.Text = "Your request is processed"
				}

			} else {
				if db.IsAdmin(chatID) && isBroadcastMode[chatID] {
					functions.HandleAdminBroadcast(bot, update.Message, update, db, telegramChannel, &isBroadcastMode)
					continue
				}

				switch update.Message.Command() {
				case "start":
					msg.Text = "Hello, HSE Student!"
				case "help":
					msg.Text = "Usage: /start, /help, /broadcast"
				case "broadcast":
					if db.IsAdmin(chatID) {
						msg.Text = "Please enter the subject and control element, e.g., 'Algebra lecture 2'."
						isBroadcastMode[chatID] = true
					} else {
						msg.Text = "You are not an admin"
					}
				case "get_materials_search":
					materialStep[chatID] = "awaiting_subject"
					msg.Text = "Please enter the subject name"
				case "delete_subscriber":
					if chatID == adminMain {
						msg.Text = "Send subscriber's ID"
						isDeleteSubscriberMode = true
					}
				case "delete_admin":
					if chatID == adminMain {
						msg.Text = "Send admin's ID"
						isDeleteAdminMode = true
					}
				case "become_admin":
					if db.IsAdmin(chatID) {
						msg.Text = "You are already an admin"
					} else {
						if !inProcessAdminReq[chatID] {
							msg.Text = "Your admin request is processed"
							inProcessAdminReq[chatID] = true
							admins.SendAdminRequest(bot, chatID, adminMain, db, update.Message.Chat)
						} else {
							msg.Text = "Your admin request already is processed"
						}
					}
				case "requests":
					subscribers.HandleSubscriberRequests(bot, chatID, db)
				case "admin_requests":
					admins.HandleAdminRequests(bot, update.Message, chatID, db)
				case "get_admins_info":
					admins.HandleGetAdminsInfo(bot, chatID, db)
				case "get_materials":
					functions.HandleGetSubjects(bot, update, chatID, db, 0)
				case "delete_material":
					if chatID == adminMain {
						msg.Text = "Send subject, control element, element number (ex. 'Алгебра КР 1')"
						isDeleteMaterialMode = true
					}
				case "delete_subject":
					if chatID == adminMain {
						msg.Text = "Send subject name"
						isDeleteSubjectMode = true
					}
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
								functions.SendMaterialSearch(bot, chatID, db)
							}
						}
					}
					if isDeleteSubscriberMode && chatID == adminMain {
						deleteID, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
						if err != nil {
							log.Printf("Error converting delete_subscriber_id to int: %v\n", err)
							msg.Text = "There isn't this subscriber"
						} else {
							if subscribers.DeleteSubscriber(bot, int64(deleteID), adminMain, db) {
								msg.Text = "Subscriber was deleted"
							} else {
								msg.Text = "We can't delete this subscriber"
							}
							isDeleteSubscriberMode = false
						}

					}
					if isDeleteAdminMode && chatID == adminMain {
						deleteID, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
						if err != nil {
							log.Printf("Error converting delete_admin_id to int: %v\n", err)
							msg.Text = "There isn't this admin"
						} else {
							if admins.DeleteAdmin(bot, int64(deleteID), adminMain, db) {
								msg.Text = "Admin was deleted"
							} else {
								msg.Text = "We can't delete this admin"
							}
							isDeleteAdminMode = false
						}

					}
					if isDeleteMaterialMode && chatID == adminMain {
						parts := strings.Split(strings.TrimSpace(update.Message.Text), " ")
						if len(parts) != 3 {
							log.Printf("Error converting parts count: %v\n", err)
							isDeleteMaterialMode = false
							msg := tgbotapi.NewMessage(chatID, "Неверное название для удаления материала.")
							if _, err := bot.Send(msg); err != nil {
								log.Printf("Error sending message: %v\n", err)
							}
						}
						number, err := strconv.Atoi(parts[2])
						if err != nil {
							log.Printf("Error converting parts number: %v\n", err)
							msg := tgbotapi.NewMessage(chatID, "Неверный номер материала для удаления.")
							isDeleteMaterialMode = false
							if _, err := bot.Send(msg); err != nil {
								log.Printf("Error sending message: %v\n", err)
							}
						}
						if !db.IsMaterialExists(parts[0], parts[1], number) {
							msg := tgbotapi.NewMessage(chatID, "Не удалось найти материал для удаления")
							if _, err = bot.Send(msg); err != nil {
								log.Printf("Error sending message: %v\n", err)
							}
						}
						err = db.RemoveMaterial(parts[0], parts[1], number)
						if err != nil {
							log.Printf("Error removing material: %v\n", err)
							msg := tgbotapi.NewMessage(chatID, "Ошибка при удалении материала.")
							if _, err := bot.Send(msg); err != nil {
								log.Printf("Error sending message: %v\n", err)
							}
						} else {
							log.Printf("Removed material: %v\n", update.Message.Text)
							msg := tgbotapi.NewMessage(chatID, "")
							msg.Text = fmt.Sprintf("Материал '%v' удалён.", update.Message.Text)
							if _, err := bot.Send(msg); err != nil {
								log.Printf("Error sending message: %v\n", err)
							}

							if db.CountMaterialForSubject(parts[0]) == 0 {
								log.Printf("Subject doen't exist after removing material: %v\n", parts[0])
							}
						}
					}
					if isDeleteSubjectMode && chatID == adminMain {
						subject := update.Message.Text
						exists, err := db.SubjectExists(subject)
						if err != nil {
							log.Printf("Error checking if subject exists: %v\n", err)
							msg := tgbotapi.NewMessage(chatID, "Ошибка при проверке предмета")
							if _, err := bot.Send(msg); err != nil {
								log.Printf("Error sending message: %v\n", err)
							}
							isDeleteSubjectMode = false
						} else {
							if !exists {
								log.Printf("Subject doesn't exist: %v\n", err)
								msg := tgbotapi.NewMessage(chatID, "Предмет не существует")
								if _, err := bot.Send(msg); err != nil {
									log.Printf("Error sending message: %v\n", err)
								}
								isDeleteSubjectMode = false
							} else {
								err = db.DeleteSubject(subject)
								if err != nil {
									log.Printf("Error removing subject: %v\n", err)
									msg := tgbotapi.NewMessage(chatID, "Ошибка при удалении предмета")
									if _, err := bot.Send(msg); err != nil {
										log.Printf("Error sending message: %v\n", err)
									}
								} else {
									log.Printf("Removed subject: %v\n", subject)
									msg := tgbotapi.NewMessage(chatID, "")
									msg.Text = fmt.Sprintf("Предмет '%v' удалён", subject)
									if _, err := bot.Send(msg); err != nil {
										log.Printf("Error sending message: %v\n", err)
									}
								}
								isDeleteSubjectMode = false
							}
						}
					}

				}
			}

			if msg.Text != "" {
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Send message error to %v: %v", chatID, err)
				}
			}

		} else if update.CallbackQuery != nil && db.IsAdmin(update.CallbackQuery.From.ID) {
			callbackData := update.CallbackQuery.Data

			if strings.HasPrefix(callbackData, "request_admin_") ||
				strings.HasPrefix(callbackData, "accept_admin_") ||
				strings.HasPrefix(callbackData, "reject_admin_") {
				admins.HandleAdminRequestCallback(bot, update.CallbackQuery, db, &inProcessAdminReq)
			} else if strings.HasPrefix(callbackData, "request_subscriber_") ||
				strings.HasPrefix(callbackData, "accept_subscriber_") ||
				strings.HasPrefix(callbackData, "reject_subscriber_") {
				subscribers.HandleRequestCallback(bot, update.CallbackQuery, db, &inProcessSubReq)
			} else if strings.HasPrefix(callbackData, "get_admin_info_") {
				admins.HandleAdminInfoCallback(bot, update.CallbackQuery, db)
			} else {
				functions.HandleCallbackQuery(bot, update, db, telegramChannel, &isBroadcastMode)
			}
		} else if update.CallbackQuery != nil && db.IsSubscriber(update.CallbackQuery.From.ID) {
			functions.HandleCallbackQuery(bot, update, db, telegramChannel, &isBroadcastMode)
		}
	}
}
