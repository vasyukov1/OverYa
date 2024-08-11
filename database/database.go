package database

import (
	"database/sql"
	"fmt"
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

	//createSubjectsTable := `
	//CREATE TABLE IF NOT EXISTS Admins (
	//    ID SERIAL PRIMARY KEY,
	//    name VARCHAR(255) NOT NULL
	//);`
	//
	//createElementsTable := `
	//CREATE TABLE IF NOT EXISTS Elements (
	//    ID SERIAL PRIMARY KEY,
	//    name VARCHAR(255) NOT NULL,
	//    type VARCHAR(50) NOT NULL
	//);`
	//
	//createSubjectElementsTable := `
	//CREATE TABLE IF NOT EXISTS SubjectElements (
	//    ID SERIAL PRIMARY KEY,
	//    subject_id INT NOT NULL REFERENCES Subjects(ID),
	//    element_id INT NOT NULL REFERENCES Elements(ID)
	//);`

	queries := []string{
		createSubscribersTable,
		createAdminsTable,
		//createSubjectsTable,
		//createElementsTable,
		//createSubjectElementsTable,
	}
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("Error executing SQL request: %v", err)
		}
	}
	return nil
}

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

func (db *DB) GetSubscribers() map[int64]bool {
	subscribers := make(map[int64]bool)
	rows, err := db.Query("SELECT ID FROM Subscribers")
	if err != nil {
		log.Fatalf("Failed to query subscribers: %v", err)
	}
	defer rows.Close()

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
