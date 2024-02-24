package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// HoneyTokensDatabase is where the configured honeytokens are stored
var HoneyTokensDatabase *gorm.DB

// Honeytoken struct represents the database structure for the honeytokens
type Honeytoken struct {
	ID             uint   `json:"id" gorm:"primary_key"`
	URL            string `json:"url"`
	Title          string `json:"title"`
	Comment        string `json:"comment"`
	NotifyReceiver string `json:"notifyReceiver"`
}

// ConnectDatabase opens the database for requests
func ConnectDatabase(file string) error {
	// init Non-Default Logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level, GORM defined log levels: Silent, Error, Warn, Info
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)

	// open (or create if not exists) DB
	database, err := gorm.Open(sqlite.Open(file), &gorm.Config{Logger: newLogger})
	if err != nil {
		return err
	}

	err = database.AutoMigrate(&Honeytoken{})
	if err != nil {
		return err
	}

	HoneyTokensDatabase = database
	return nil
}

// Find searches for an url and returns an Honeytoken if found
func Find(url string) (Honeytoken, error) {
	var tokenFound Honeytoken
	result := HoneyTokensDatabase.Find(&tokenFound, "url = ?", url)
	if result.Error != nil {
		return Honeytoken{}, result.Error
	}

	if tokenFound.ID == 0 {
		return Honeytoken{}, errors.New("no token found")
	}

	fmt.Println("Found entry:", tokenFound)
	return tokenFound, nil
}
