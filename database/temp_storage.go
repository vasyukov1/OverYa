package database

// ----------------------- TEMP STORAGE -----------------------

var tempSubject = make(map[int64]string)
var tempControlElement = make(map[int64]string)
var tempElementNumber = make(map[int64]string)

func (db *DB) SetTempSubject(chatID int64, subject string) {
	tempSubject[chatID] = subject
}

func (db *DB) SetTempControlElement(chatID int64, controlElement string) {
	tempControlElement[chatID] = controlElement
}

func (db *DB) SetTempElementNumber(chatID int64, elementNumber string) {
	tempElementNumber[chatID] = elementNumber
}

func (db *DB) GetTempSubject(chatID int64) string {
	return tempSubject[chatID]
}

func (db *DB) GetTempControlElement(chatID int64) string {
	return tempControlElement[chatID]
}

func (db *DB) GetTempElementNumber(chatID int64) string {
	return tempElementNumber[chatID]
}
