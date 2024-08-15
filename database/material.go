package database

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

// ----------------------- MATERIALS -----------------------

func (db *DB) AddMaterial(subjectName string, controlElement string, elementNumber int, fileIDs []string, description string) error {
	_, err := db.Exec(
		"INSERT INTO Materials (SubjectName, ControlElement, ElementNumber, FileIDs, Description) "+
			"VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING",
		subjectName, controlElement, elementNumber, pq.Array(fileIDs), description)
	return err
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
