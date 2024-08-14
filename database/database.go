package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"log"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "alexvasyukov"
	password = "123"
	dbname   = "postgres"
)

type DB struct {
	*sql.DB
}

func NewDB() (*DB, error) {
	dbInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) CreateTables() error {
	createSubscribersTable := `
    CREATE TABLE IF NOT EXISTS Subscribers (
        ID SERIAL PRIMARY KEY
    );`

	createAdminsTable := `
	CREATE TABLE IF NOT EXISTS Admins (
	    ID SERIAL PRIMARY KEY,
	    count_of_posts INT NOT NULL DEFAULT 0
	);`

	createSubjectsTable := `
	CREATE TABLE IF NOT EXISTS Subjects (
	  Name TEXT NOT NULL PRIMARY KEY
	);`

	createMaterialsTable := `
	CREATE TABLE IF NOT EXISTS Materials (
		SubjectName TEXT REFERENCES Subjects(Name) ON DELETE CASCADE,
	    ControlElement TEXT NOT NULL,
	    ElementNumber TEXT NOT NULL,
	    FileIDs TEXT[] NOT NULL,
	    Description TEXT,
	    PRIMARY KEY (SubjectName, ControlElement, ElementNumber)
	);`

	createSubscriberRequestsTable := `
	CREATE TABLE IF NOT EXISTS subscriber_requests (
	    id SERIAL PRIMARY KEY,
	    first_name TEXT,
	    last_name TEXT,
	    user_name TEXT
	);`

	queries := []string{
		createSubscribersTable,
		createAdminsTable,
		createSubjectsTable,
		createMaterialsTable,
		createSubscriberRequestsTable,
	}
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("Error executing SQL request: %v", err)
			return err
		}
	}
	return nil
}

// ----------------------- USERS -----------------------

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

// DeleteSubscriber Need checking
func (db *DB) DeleteSubscriber(chatID int64) {
	_, err := db.Exec("DELETE FROM Subscribers WHERE id=$1", chatID)
	if err != nil {
		log.Printf("Delete Subscribers error: %v", err)
	}
}

// AddAdmin Need checking
func (db *DB) AddAdmin(chatID int64) {
	_, err := db.Exec("INSERT INTO Admins (ID) VALUES ($1) ON CONFLICT DO NOTHING", chatID)
	if err != nil {
		log.Printf("Add Subscribers error: %v", err)
	}
}

// IsAdmin Need checking
func (db *DB) IsAdmin(chatID int64) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM Admins WHERE id=$1)", chatID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking subscriber existence: %v", err)
		return false
	}
	return exists
}

// DeleteAdmin Need checking
func (db *DB) DeleteAdmin(chatID int64) {
	_, err := db.Exec("DELETE FROM Admins WHERE id=$1", chatID)
	if err != nil {
		log.Printf("Delete Admin error: %v", err)
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

// ----------------------- REQUEST ------------------------

// GetSubscriberRequests NEED CHECKING
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

// CountSubscriberRequest NEED CHECKING
func (db *DB) CountSubscriberRequest() int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM subscriber_requests").Scan(&count)
	if err != nil {
		log.Printf("Failed to count subscriber requests: %v", err)
		return 0
	}
	return count
}

func (db *DB) AddSubscriberRequest(chatID int64, firstName, lastName, userName string) error {
	//_, err := db.Exec("INSERT INTO SubscriberRequests (ID) VALUES ($1) ON CONFLICT DO NOTHING", chatID)
	_, err := db.Exec(`
		INSERT INTO subscriber_requests (id, first_name, last_name, user_name)
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

// ----------------------- SUBJECTS -----------------------

func (db *DB) SubjectExists(subjectName string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM Subjects WHERE Name=$1)", subjectName).Scan(&exists)
	return exists, err
}

func (db *DB) AddSubject(subjectName string) error {
	_, err := db.Exec("INSERT INTO Subjects (Name) VALUES ($1) ON CONFLICT DO NOTHING", subjectName)
	return err
}

// ----------------------- MATERIALS -----------------------

func (db *DB) AddMaterial(subjectName string, controlElement string, elementNumber int, fileIDs []string, description string) error {
	_, err := db.Exec(
		"INSERT INTO Materials (SubjectName, ControlElement, ElementNumber, FileIDs, Description) "+
			"VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING",
		subjectName, controlElement, elementNumber, pq.Array(fileIDs), description)
	return err
}

func (db *DB) GetMaterial(chatID int64) ([]string, string, error) {
	var fileIDs []string
	var description string
	err := db.QueryRow(
		"SELECT FileIDs, Description FROM Materials WHERE SubjectName = $1 AND ControlElement = $2 AND ElementNumber = $3",
		tempSubject[chatID], tempControlElement[chatID], tempElementNumber[chatID]).Scan(pq.Array(&fileIDs), &description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", err
	}
	return fileIDs, description, nil
}

// ----------------------- TEMP STORAGE -----------------------

var tempSubject = make(map[int64]string)
var tempControlElement = make(map[int64]string)
var tempElementNumber = make(map[int64]int)

func (db *DB) SetTempSubject(chatID int64, subject string) {
	tempSubject[chatID] = subject
}

func (db *DB) SetTempControlElement(chatID int64, controlElement string) {
	tempControlElement[chatID] = controlElement
}

func (db *DB) SetTempElementNumber(chatID int64, elementNumber int) {
	tempElementNumber[chatID] = elementNumber
}

func (db *DB) GetTempSubject(chatID int64) string {
	return tempSubject[chatID]
}

func (db *DB) GetTempControlElement(chatID int64) string {
	return tempControlElement[chatID]
}

func (db *DB) GetTempElementNUmber(chatID int64) int {
	return tempElementNumber[chatID]
}
