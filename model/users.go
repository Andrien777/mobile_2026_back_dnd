package model

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

type InternalUser struct {
	gorm.Model
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
}

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
	db.Debug().AutoMigrate(&InternalUser{})
}

func GetDB() *gorm.DB {
	return db
}

func Check_account(user InternalUser) bool {
	if len(user.Password) < 6 {
		return false
	}

	temp := &InternalUser{}

	//проверка на наличие ошибок и дубликатов электронных писем
	err := GetDB().Table("accounts").Where("Username = ?", user.Username).First(temp).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	if temp.Username != "" {
		return false
	}

	return true
}
