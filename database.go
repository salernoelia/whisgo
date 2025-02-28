package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Transcription struct {
	Id		int
	Text	string
	Timestamp	time.Time
} 



const (
	dbFilename = "whisgo.db"
)

var (
	db       *sql.DB
	dbOnce   sync.Once
	dbPath string
)


func init() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Error getting user config dir:", err)
		configDir = "."
	}
	dbPath = filepath.Join(configDir, "whisgo", dbFilename)
}

func GetDB() (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		dir := filepath.Dir(dbPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Printf("Failed to create database directory: %v\n", err)
				return
			}
		}

		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			fmt.Printf("Failed to open database: %v\n", err)
			return
		}

		err = createTranscriptionsTable(db)
		if err != nil {
			fmt.Printf("Failed to create transcriptions table: %v\n", err)
			db = nil
			return
		}
	})

	if db == nil && err != nil {
		return nil, err
	}

	return db, nil
}

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
