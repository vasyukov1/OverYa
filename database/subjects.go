package database

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
