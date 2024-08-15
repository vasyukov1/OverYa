package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"log"
	"slices"
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

	createSubscriberRequestsTable := `
	CREATE TABLE IF NOT EXISTS subscriber_requests (
	    ID SERIAL PRIMARY KEY,
	    first_name TEXT,
	    last_name TEXT,
	    user_name TEXT
	);`

	createAdminsTable := `
	CREATE TABLE IF NOT EXISTS admins (
	    ID SERIAL PRIMARY KEY,
	    first_name TEXT,
	    last_name TEXT,
	    user_name TEXT
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

	createAdminRequestsTable := `
	CREATE TABLE IF NOT EXISTS admin_requests (
	    ID SERIAL PRIMARY KEY,
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
		createAdminRequestsTable,
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

// -------------------- SUBJECTS --------------------

func (db *DB) GetSubjects() []string {

	var subjects []string
	rows, err := db.Query("SELECT Name FROM Subjects")
	if err != nil {
		log.Printf("Failed to query subjects request: %v", err)
		return nil
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			log.Fatalf("Failed to close rows: %v", err)
		}
	}(rows)
	for rows.Next() {
		var subject string
		if err = rows.Scan(&subject); err != nil {
			log.Printf("Failed to scan subject: %v", err)
		}
		subjects = append(subjects, subject)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating subjects requests: %v", err)
	}
	return subjects
}

func (db *DB) GetControlElements(subject string) []string {
	query := `
		SELECT DISTINCT ControlElement 
		FROM Materials 
		WHERE SubjectName = $1
	`
	rows, err := db.Query(query, subject)
	if err != nil {
		log.Printf("Error querying control elements: %v\n", err)
		return nil
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var controlElements []string
	for rows.Next() {
		var controlElement string
		if err = rows.Scan(&controlElement); err != nil {
			log.Printf("Error scanning control element: %v\n", err)
			continue
		}
		if !slices.Contains(controlElements, controlElement) {
			controlElements = append(controlElements, controlElement)
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error in control element rows iteration: %v\n", err)
	}

	return controlElements
}

func (db *DB) GetElementNumber(subject string, controlElement string) []int {
	query := `
		SELECT DISTINCT ElementNumber 
		FROM Materials 
		WHERE SubjectName = $1 AND ControlElement = $2
	`
	rows, err := db.Query(query, subject, controlElement)
	if err != nil {
		log.Printf("Error querying element number: %v\n", err)
		return nil
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var numbers []int
	for rows.Next() {
		var number int
		if err = rows.Scan(&number); err != nil {
			log.Printf("Error scanning element number: %v\n", err)
			continue
		}
		numbers = append(numbers, number)

	}

	if err := rows.Err(); err != nil {
		log.Printf("Error in element number rows iteration: %v\n", err)
	}

	return numbers
}

func (db *DB) GetMaterial(subject string, controlElement string, elementNumber int) ([]string, string, error) {
	var fileIDs []string
	var description string

	query := `SELECT FileIDs, Description FROM Materials WHERE SubjectName = $1 AND ControlElement = $2 AND ElementNumber = $3`
	err := db.QueryRow(query, subject, controlElement, elementNumber).Scan(pq.Array(&fileIDs), &description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", err
		}
		return nil, "", err
	}

	return fileIDs, description, nil
}
