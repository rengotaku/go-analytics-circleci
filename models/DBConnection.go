package models

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm/logger"

	"gorm.io/gorm"
)

var (
	dbFileName = ""
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dbFileName = dir + "/db/circleci.sqlite3"
	log.Printf(dbFileName)
}

func InitMigration() error {
	db, err := gorm.Open(sqlite.Open(dbFileName), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&Pipeline{})
	db.AutoMigrate(&Workflow{})
	db.AutoMigrate(&WorkflowJob{})
	db.AutoMigrate(&Job{})

	return nil
}

func Connection() (*gorm.DB, error) {
	logLevel := logger.Info

	newLogger := logger.New(
		log.New(os.Stdout, "\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logLevel,    // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log
			Colorful:                  true,        // Disable color
		},
	)

	db, err := gorm.Open(sqlite.Open(dbFileName), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}
