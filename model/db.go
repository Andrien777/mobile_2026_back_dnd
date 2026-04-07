package model

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func init() {
	dbHost := "postgres"
	username := "user"
	password := "password"
	dbName := "dnd_back"
	dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username, dbName, password)

	var (
		conn *gorm.DB
		err  error
	)

	for attempt := 1; attempt <= 30; attempt++ {
		conn, err = gorm.Open("postgres", dbUri)
		if err == nil {
			sqlDB := conn.DB()
			if pingErr := sqlDB.Ping(); pingErr == nil {
				break
			} else {
				err = pingErr
				_ = sqlDB.Close()
			}
		}
		log.Printf("waiting for postgres (%d/30): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil || conn == nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	db = conn
	if err := db.AutoMigrate(&InternalUser{}, &InternalCharacter{}, &InternalCharacterShare{}, &InternalSpell{}, &InternalItem{}, &InternalFeat{}).Error; err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_character_share_unique ON internal_character_shares (character_id, shared_with_username)").Error; err != nil {
		log.Fatalf("failed to add unique index for character shares: %v", err)
	}
}

func GetDB() *gorm.DB {
	return db
}
