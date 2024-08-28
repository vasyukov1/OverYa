package database

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"log"
	"slices"
)

// ----------------------- MATERIALS -----------------------

func (db *DB) AddMaterial(subjectName string, controlElement string, elementNumber int, fileIDs []string, description string) error {
	query := "INSERT INTO Materials (SubjectName, ControlElement, ElementNumber, FileIDs, Description) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING"
	_, err := db.Exec(query, subjectName, controlElement, elementNumber, pq.Array(fileIDs), description)
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

// -------------------- EDITION --------------------

// EditName has some problems ??
func (db *DB) EditName(subject string, controlElement string, number int, old []string) error {
	query := `
		UPDATE Materials
		SET SubjectName = $1, ControlElement = $2, ElementNumber = $3
		WHERE SubjectName = $4 AND ControlElement = $5 AND ElementNumber = $6
	`
	_, err := db.Exec(query, subject, controlElement, number, old[0], old[1], old[2])
	if err != nil {
		log.Println("Ошибка выполнения запроса:", err)
		return err
	}
	return nil
}
