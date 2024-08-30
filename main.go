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
				if !functions.InProcessSubReq[chatID] {
					msg.Text = "Чтобы использовать бота, станьте подписчиком /become_subscriber"
					switch update.Message.Command() {
					case "become_subscriber":
						msg.Text = "Ваша заявка обрабатывается"
						functions.InProcessSubReq[chatID] = true
						subscribers.SendSubscribeRequest(bot, chatID, adminMain, db, update.Message.Chat)
					}
				} else {
					msg.Text = "Ваша заявка обрабатывается"
				}

			} else {
				if db.IsAdmin(chatID) && functions.IsBroadcastMode[chatID] {
					functions.HandleAdminBroadcast(bot, update.Message, update, db, telegramChannel, &functions.IsBroadcastMode)
					continue
				}
				if db.IsAdmin(chatID) && functions.IsNewsMode[chatID] {
					functions.News(bot, chatID, update.Message.Text, db)
					functions.IsNewsMode[chatID] = false
					continue
				}

				switch update.Message.Command() {
				case "start":
					msg.Text = "Привет, студент!\n\nНажми /help, чтобы ознакомиться с функционалом"
				case "help":
					if chatID == adminMain {
						msg.Text = fmt.Sprintf(
							`
/start - начало работы
/become_admin - стать админом
/get_materials - выбрать материал
/get_materials_search - найти материал по названию
/count_subscribers - количество подписчиков
/count_admins - количество админов
/broadcast - отправить материал
/news - отправить новость
/delete_material - удалить материал
/delete_subject - удалить предмет
/delete_subscriber - удалить подписчика
/delete_admin - удалить админа
/subscribers - список подписчиков
/admins - список админов
/requests_subscriber - заявки на подписку
/requests_admin - заявки на админство
`)
					} else if db.IsAdmin(chatID) {
						msg.Text = fmt.Sprintf(
							`
/start - начало работы
/become_admin - стать админом
/get_materials - выбрать материал
/get_materials_search - найти материал по названию
/count_subscribers - количество подписчиков
/count_admins - количество админов
/broadcast - отправить материал
/news - отправить новость
/delete_material - удалить материал
/delete_subject - удалить предмет
`)
					} else {
						msg.Text = fmt.Sprintf(
							`
/start - начало работы
/become_admin - стать админом
/get_materials - выбрать материал
/get_materials_search - найти материал по названию
/count_subscribers - количество подписчиков
/count_admins - количество админов
`)
					}
				case "go_main":
					functions.GoToMain(chatID, db, bot)
				case "broadcast":
					if db.IsAdmin(chatID) {
						msg.Text = "Введите название предмета, элемент контроля и номер строго в формате 'Матан лекция 2'."
						functions.IsBroadcastMode[chatID] = true
					} else {
						msg.Text = "Вы не админ"
					}
				case "get_materials_search":
					functions.MaterialStep[chatID] = "awaiting_subject"
					msg.Text = "Введите название предмета"
				case "delete_subscriber":
					if chatID == adminMain {
						msg.Text = "Отправьте ID подписчика"
						functions.IsDeleteSubscriberMode = true
					}
				case "delete_admin":
					if chatID == adminMain {
						msg.Text = "Отправьте ID админа"
						functions.IsDeleteAdminMode = true
					}
				case "become_admin":
					if db.IsAdmin(chatID) {
						msg.Text = "Вы уже админ"
					} else {
						if !functions.InProcessAdminReq[chatID] {
							msg.Text = "Ваша заявка на админа рассматривается"
							functions.InProcessAdminReq[chatID] = true
							admins.SendAdminRequest(bot, chatID, adminMain, db, update.Message.Chat)
						} else {
							msg.Text = "Ваша заявка на админа уже рассматривается"
						}
					}
				case "subscribers":
					if chatID == adminMain {
						subscribers.GetSubscribers(bot, update, chatID, db, 0)
					} else {
						msg.Text = "Вы не админ"
					}
				case "admins":
					if chatID == adminMain {
						admins.GetAdmins(bot, update, chatID, db, 0)
					} else {
						msg.Text = "Вы не админ"
					}
				case "requests_subscriber":
					subscribers.HandleSubscriberRequests(bot, update, chatID, db, 0)
				case "requests_admin":
					admins.HandleAdminRequests(bot, update, chatID, db, 0)
				case "get_materials":
					functions.HandleGetSubjects(bot, update, chatID, db, 0)
				case "delete_material":
					if db.IsAdmin(chatID) {
						msg.Text = "Отправьте название предмета, элемнт контроля и номер строго по формату (например, Алгебра КР 1)"
						functions.IsDeleteMaterialMode = true
					}
				case "delete_subject":
					if chatID == adminMain {
						msg.Text = "Отправьте название предмета"
						functions.IsDeleteSubjectMode = true
					}
				case "count_subscribers":
					countSubscribers := db.CountSubscribers()
					msg.Text = fmt.Sprintf("Всего %v подписчиков!", countSubscribers)
				case "count_admins":
					countAdmins := db.CountAdmins()
					msg.Text = fmt.Sprintf("Всего %v админов!", countAdmins)
				case "news":
					msg.Text = "Напишите новость"
					functions.IsNewsMode[chatID] = true
				default:
					msg.Text = "Команда на найдена. Используйте /help"
				}

				if update.Message.Command() == "" {

					if functions.MaterialStep[chatID] != "" {
						switch functions.MaterialStep[chatID] {
						case "awaiting_subject":
							msg.Text = "Введите элемент контроля (например, семинар)"
							db.SetTempSubject(chatID, update.Message.Text)
							functions.MaterialStep[chatID] = "awaiting_control_element"
						case "awaiting_control_element":
							msg.Text = "Введите номер элемента контроля"
							db.SetTempControlElement(chatID, update.Message.Text)
							functions.MaterialStep[chatID] = "awaiting_element_number"
						case "awaiting_element_number":
							elementNumberForGet := update.Message.Text
							db.SetTempElementNumber(chatID, elementNumberForGet)
							functions.MaterialStep[chatID] = ""
							subject := db.GetTempSubject(chatID)
							controlElement := db.GetTempControlElement(chatID)
							number := db.GetTempElementNumber(chatID)
							functions.SendMaterial(bot, chatID, db, subject, controlElement, number)
						}
					}

					if functions.IsDeleteSubscriberMode && chatID == adminMain {
						deleteID, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
						if err != nil {
							log.Printf("Error converting delete_subscriber_id to int: %v\n", err)
							msg.Text = "Такого подписчика нет"
						} else {
							if subscribers.DeleteSubscriber(bot, int64(deleteID), adminMain, db) {
								msg.Text = "Подписчик был удалён"
							} else {
								msg.Text = "Я не смог удалить подписчика"
							}
							functions.IsDeleteSubscriberMode = false
						}
						continue
					}
					if functions.IsDeleteAdminMode && chatID == adminMain {
						deleteID, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
						if err != nil {
							log.Printf("Error converting delete_admin_id to int: %v\n", err)
							msg.Text = "Такого админа нет"
						} else {
							if admins.DeleteAdmin(bot, int64(deleteID), adminMain, db) {
								msg.Text = "Админ был удалён"
							} else {
								msg.Text = "Я не смог удалить админа"
							}
							functions.IsDeleteAdminMode = false
						}

					}
					if functions.IsDeleteMaterialMode && chatID == adminMain {
						parts := strings.Split(strings.TrimSpace(update.Message.Text), " ")
						if len(parts) != 3 {
							log.Printf("Error converting parts count: %v\n", err)
							functions.IsDeleteMaterialMode = false
							msg := tgbotapi.NewMessage(chatID, "Неверное название материала для удаления")
							if _, err := bot.Send(msg); err != nil {
								log.Printf("Error sending message: %v\n", err)
							}
						} else {
							if !db.IsMaterialExists(parts[0], parts[1], parts[2]) {
								msg := tgbotapi.NewMessage(chatID, "Не удалось найти материал для удаления")
								if _, err = bot.Send(msg); err != nil {
									log.Printf("Error sending message: %v\n", err)
								}
							}
							functions.IsDeleteMaterialMode = false
							err = db.RemoveMaterial(parts[0], parts[1], parts[2])
							if err != nil {
								log.Printf("Error removing material: %v\n", err)
								msg := tgbotapi.NewMessage(chatID, "Ошибка при удалении материала")
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
									if err := db.DeleteSubject(parts[0]); err != nil {
										log.Printf("Error removing subject: %v\n", err)
									} else {
										log.Printf("Subject doesn't exist after removing material: %v\n", parts[0])
									}
								}
							}
						}
						continue
					}
					if functions.IsDeleteSubjectMode && chatID == adminMain {
						subject := update.Message.Text
						exists, err := db.SubjectExists(subject)
						functions.IsDeleteSubjectMode = false
						if err != nil {
							log.Printf("Error checking if subject exists: %v\n", err)
							msg := tgbotapi.NewMessage(chatID, "Ошибка при проверке существования предмета")
							if _, err := bot.Send(msg); err != nil {
								log.Printf("Error sending message: %v\n", err)
							}
						} else {
							if !exists {
								log.Printf("Subject doesn't exist: %v\n", subject)
								msg := tgbotapi.NewMessage(chatID, "Предмет не существует")
								if _, err := bot.Send(msg); err != nil {
									log.Printf("Error sending message: %v\n", err)
								}
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
							}
						}
						continue
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
				admins.HandleAdminRequestCallback(bot, update.CallbackQuery, db, &functions.InProcessAdminReq)
			} else if strings.HasPrefix(callbackData, "request_subscriber_") ||
				strings.HasPrefix(callbackData, "accept_subscriber_") ||
				strings.HasPrefix(callbackData, "reject_subscriber_") {
				subscribers.HandleRequestCallback(bot, update.CallbackQuery, db, &functions.InProcessSubReq)
			} else if strings.HasPrefix(callbackData, "get_admin_info_") {
				admins.HandleAdminInfoCallback(bot, update.CallbackQuery, db)
			} else {
				functions.HandleCallbackQuery(bot, update, db, telegramChannel, &functions.IsBroadcastMode)
			}
		} else if update.CallbackQuery != nil && db.IsSubscriber(update.CallbackQuery.From.ID) {
			functions.HandleCallbackQuery(bot, update, db, telegramChannel, &functions.IsBroadcastMode)
		}
	}
}
