package database

import (
	"log"
)

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
