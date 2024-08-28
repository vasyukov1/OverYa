package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

func (db *DB) IsAdmin(chatID int64) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM admins WHERE id=$1)", chatID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking subscriber existence: %v", err)
		return false
	}
	return exists
}

func (db *DB) AddAdmin(chatID int64) {
	_, err := db.Exec("INSERT INTO admins (ID) VALUES ($1) ON CONFLICT DO NOTHING", chatID)
	if err != nil {
		log.Printf("Add Subscribers error: %v", err)
	}
}

func (db *DB) DeleteAdmin(chatID int64) {
	_, err := db.Exec("DELETE FROM admins WHERE id=$1", chatID)
	if err != nil {
		log.Printf("Delete Admin error: %v", err)
	}
}

func (db *DB) GetAdmins() map[int64]bool {
	admins := make(map[int64]bool)
	rows, err := db.Query("SELECT ID FROM admins")
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
		admins[chatID] = true
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating subscribers: %v", err)
	}
	return admins
}

func (db *DB) GetAdminInfo(adminID int64) (string, error) {
	var firstName, lastName, userName string
	query := `SELECT first_name, last_name, user_name FROM admins WHERE id = $1`
	err := db.QueryRow(query, adminID).Scan(&firstName, &lastName, &userName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("admin with ID %d not found", adminID)
		}
		return "", err
	}
	userInfo := fmt.Sprintf("%s %s", firstName, lastName)
	if userName != "" {
		userInfo = fmt.Sprintf("%s [@%s]", userInfo, userName)
	}
	return userInfo, nil
}

func (db *DB) CountAdmins() int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM admins").Scan(&count)
	if err != nil {
		log.Printf("Failed to count admins: %v", err)
		return 0
	}
	return count
}

// ----------------------- ADMIN REQUEST ------------------------

func (db *DB) AddAdminRequest(chatID int64, firstName, lastName, userName string) error {
	_, err := db.Exec(`
		INSERT INTO admin_requests (ID, first_name, last_name, user_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) 
		DO UPDATE SET 
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			user_name = EXCLUDED.user_name
	`, chatID, firstName, lastName, userName)
	return err
}

func (db *DB) DeleteAdminRequest(chatID int64) error {
	_, err := db.Exec("DELETE FROM admin_requests WHERE id=$1", chatID)
	return err
}

func (db *DB) GetAdminRequests() ([]int64, error) {
	var requests []int64
	rows, err := db.Query("SELECT ID FROM admin_requests")
	if err != nil {
		log.Fatalf("Failed to query admin request: %v", err)
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
			log.Printf("Failed to scan admin request ID %v: %s", chatID, err)
			continue
		}
		requests = append(requests, chatID)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating admin requests: %v", err)
	}
	return requests, nil
}

func (db *DB) GetAdminRequestInfo(requestID int64) (string, error) {
	var firstName, lastName, userName string
	query := `SELECT first_name, last_name, user_name FROM admin_requests WHERE id = $1`
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

func (db *DB) CountAdminRequest() int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM admin_requests").Scan(&count)
	if err != nil {
		log.Printf("Failed to count admin requests: %v", err)
		return 0
	}
	return count
}
