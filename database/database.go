package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/vasyukov1/Overbot/config"
	"log"
)

type DB struct {
	*sql.DB
}

func NewDB() (*DB, error) {
	cfg := config.LoadConfig()
	host := cfg.Host
	port := cfg.Port
	user := cfg.User
	password := cfg.Password
	dbname := cfg.Dbname

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
