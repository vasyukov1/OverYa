package database

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"log"
)

// ----------------------- MATERIALS -----------------------

func (db *DB) AddMaterial(subjectName string, controlElement string, elementNumber int, fileIDs []string, description string) error {
	_, err := db.Exec(
		"INSERT INTO Materials (SubjectName, ControlElement, ElementNumber, FileIDs, Description) "+
			"VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING",
		subjectName, controlElement, elementNumber, pq.Array(fileIDs), description)
	return err
}

func (db *DB) RemoveMaterial(subjectName string, controlElement string, elementNumber int) error {
	query := `
		DELETE FROM Materials
		WHERE SubjectName = $1 AND ControlElement = $2 AND ElementNumber = $3;
	`
	_, err := db.Exec(query, subjectName, controlElement, elementNumber)
	if err != nil {
		log.Printf("failed to delete material: %v", err)
		return err
	}

	return nil
}

func (db *DB) RemoveMaterialBySubject(subjectName string) error {
	query := `
		DELETE FROM Materials
		WHERE SubjectName = $1;
	`
	_, err := db.Exec(query, subjectName)
	if err != nil {
		log.Printf("failed to delete material: %v", err)
		return err
	}

	return nil
}

func (db *DB) IsMaterialExists(subjectName string, controlElement string, elementNumber int) bool {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM Materials
			WHERE SubjectName = $1 AND ControlElement = $2 AND ElementNumber = $3
		);
	`
	var exists bool
	err := db.QueryRow(query, subjectName, controlElement, elementNumber).Scan(&exists)
	if err != nil {
		log.Printf("Error checking if material exists: %v", err)
		return false
	}

	return exists
}

func (db *DB) CountMaterialForSubject(subject string) int {
	query := `
		SELECT COUNT(*)
		FROM Materials
		WHERE SubjectName = $1;
	`
	var count int
	err := db.QueryRow(query, subject).Scan(&count)
	if err != nil {
		log.Printf("Error counting materials for subject: %v", err)
		return 0
	}
	return count
}

func (db *DB) GetMaterialSearch(chatID int64) ([]string, string, error) {
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
