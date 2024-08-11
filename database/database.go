package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "alex"
	password = "123"
	dbname   = "postgres"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}
	defer db.Close()

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
	CREATE TABLE IF NOT EXISTS Admins (
	    ID SERIAL PRIMARY KEY,
	    name VARCHAR(255) NOT NULL
	);`

	createElementsTable := `
    CREATE TABLE IF NOT EXISTS Elements (
        ID SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        type VARCHAR(50) NOT NULL
    );`

	createSubjectElementsTable := `
    CREATE TABLE IF NOT EXISTS SubjectElements (
        ID SERIAL PRIMARY KEY,
        subject_id INT NOT NULL REFERENCES Subjects(ID),
        element_id INT NOT NULL REFERENCES Elements(ID)
    );`

	executeSQL(db, createSubscribersTable)
	executeSQL(db, createAdminsTable)
	executeSQL(db, createSubjectsTable)
	executeSQL(db, createElementsTable)
	executeSQL(db, createSubjectElementsTable)

	fmt.Println("Database is ready")
}

func executeSQL(db *sql.DB, query string) {
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Error SQL request: %v\n", err)
	}
}
