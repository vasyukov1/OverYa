package database

import (
	"database/sql"
	"fmt"
	"log"
)

func (db *DB) IsSubscriber(chatID int64) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM Subscribers WHERE id=$1)", chatID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking subscriber existence: %v", err)
		return false
	}
	return exists
}

func (db *DB) AddSubscriber(chatID int64) {
	_, err := db.Exec("INSERT INTO Subscribers (ID) VALUES ($1) ON CONFLICT DO NOTHING", chatID)
	if err != nil {
		log.Printf("Add Subscribers error: %v", err)
	}
}

func (db *DB) DeleteSubscriber(chatID int64) {
	_, err := db.Exec("DELETE FROM Subscribers WHERE id=$1", chatID)
	if err != nil {
		log.Printf("Delete Subscribers error: %v", err)
	}
}

func (db *DB) GetSubscribers() map[int64]bool {
	subscribers := make(map[int64]bool)
	rows, err := db.Query("SELECT ID FROM Subscribers")
	if err != nil {
		log.Fatalf("Failed to query subscribers: %v", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatalf("Failed to close rows: %v", err)
		}
	}(rows)
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			log.Printf("Failed to scan subscriber ID %v: %s", chatID, err)
			continue
		}
		subscribers[chatID] = true
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating subscribers: %v", err)
	}
	return subscribers
}

// ----------------------- SUBSCRIBER REQUEST ------------------------

func (db *DB) AddSubscriberRequest(chatID int64, firstName, lastName, userName string) error {
	_, err := db.Exec(`
		INSERT INTO subscriber_requests (ID, first_name, last_name, user_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) 
		DO UPDATE SET 
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			user_name = EXCLUDED.user_name
	`, chatID, firstName, lastName, userName)
	return err
}

func (db *DB) DeleteSubscriberRequest(chatID int64) error {
	_, err := db.Exec("DELETE FROM subscriber_requests WHERE id=$1", chatID)
	return err
}

func (db *DB) GetSubscriberRequests() ([]int64, error) {
	var requests []int64
	rows, err := db.Query("SELECT ID FROM subscriber_requests")
	if err != nil {
		log.Fatalf("Failed to query subscriber request: %v", err)
		return requests, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatalf("Failed to close rows: %v", err)
		}
	}(rows)
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			log.Printf("Failed to scan subscriber request ID %v: %s", chatID, err)
			continue
		}
		requests = append(requests, chatID)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating subscriber requests: %v", err)
	}
	return requests, nil
}

func (db *DB) GetSubscriberRequestInfo(requestID int64) (string, error) {
	var firstName, lastName, userName string
	query := `SELECT first_name, last_name, user_name FROM subscriber_requests WHERE id = $1`
	err := db.QueryRow(query, requestID).Scan(&firstName, &lastName, &userName)
	if err != nil {
		return "", err
	}
	userInfo := fmt.Sprintf("%s %s", firstName, lastName)
	if userName != "" {
		userInfo = fmt.Sprintf("%s [@%s]", userInfo, userName)
	}
	return userInfo, nil
}

func (db *DB) CountSubscriberRequest() int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM subscriber_requests").Scan(&count)
	if err != nil {
		log.Printf("Failed to count subscriber requests: %v", err)
		return 0
	}
	return count
}
