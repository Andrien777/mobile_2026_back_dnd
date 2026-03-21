package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func init() {
	dbHost := "localhost"
	username := "user"
	password := "password"
	dbName := "dnd_back"
	dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username, dbName, password) //Создать строку подключения
	conn, err := gorm.Open("postgres", dbUri)
	if err != nil {
		fmt.Print(err)
	}
	db = conn
	db.Debug().AutoMigrate(&InternalUser{}, &InternalCharacter{})
}

func GetDB() *gorm.DB {
	return db
}
