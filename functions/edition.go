package functions

import (
	"github.com/vasyukov1/Overbot/database"
	"log"
)

func EditName(subject string, controlElement string, number int, old []string, db *database.DB) bool {
	if err := db.EditName(subject, controlElement, number, old); err != nil {
		log.Printf("EditName Error: %v\n", err)
		return false
	}
	oldName := old[0] + " " + old[1] + " " + old[2]
	newName := subject + " " + controlElement + " " + string(number)
	log.Printf("EditName Success: %v -> %v\n", oldName, newName)
	return true
}

func editMedia() {

}

func editDescription() {

}
