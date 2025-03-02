package main

import (
	"database/sql"
	"fmt"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func createTranscriptionsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS transcriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func AddTranscription(text string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO transcriptions (text) VALUES (?)", text)
	return err
}



func GetTranscriptions() ([]Transcription, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT id, text, timestamp FROM transcriptions ORDER BY timestamp DESC LIMIT 20")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transcriptions []Transcription
	for rows.Next() {
		var t Transcription
		err = rows.Scan(&t.Id, &t.Text, &t.Timestamp)
		if err != nil {
			return nil, err
		}
		transcriptions = append(transcriptions, t)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return transcriptions, nil
}

func ClearTranscriptions() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM transcriptions")
	return err
}


func (a *App) GetTranscriptionHistory() ([]Transcription, error) {
    return GetTranscriptions()
}

func (a *App) emitTranscriptionHistory() {
    transcriptions, err := GetTranscriptions()
    if err != nil {
        fmt.Printf("Failed to get transcriptions: %v\n", err)
        return
    }
    wailsRuntime.EventsEmit(a.ctx, "transcription-history-changed", transcriptions)
}