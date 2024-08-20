package database

import (
	"database/sql"
	"log"
)

func (db *DB) AddSubject(subjectName string) error {
	_, err := db.Exec("INSERT INTO Subjects (Name) VALUES ($1) ON CONFLICT DO NOTHING", subjectName)
	return err
}

func (db *DB) DeleteSubject(subjectName string) error {
	query := `
		DELETE FROM Subjects
		WHERE Name = $1;
	`
	_, err := db.Exec(query, subjectName)
	if err != nil {
		log.Printf("failed to delete subject: %v", err)
		return err
	}

	if db.CountMaterialForSubject(subjectName) != 0 {
		err := db.RemoveMaterialBySubject(subjectName)
		if err != nil {
			log.Printf("failed to delete all materials for subject: %v", err)
		}
	}
	return nil
}

func (db *DB) SubjectExists(subjectName string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM Subjects WHERE Name=$1)", subjectName).Scan(&exists)
	return exists, err
}

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
