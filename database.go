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
