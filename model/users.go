package model

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

type InternalUser struct {
	gorm.Model
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
}

func CheckAccount(user InternalUser) bool {
	if len(user.Password) < 6 {
		return false
	}

	temp := &InternalUser{}

	//проверка на наличие ошибок и дубликатов электронных писем
	err := GetDB().Table("internal_users").Where("Username = ?", user.Username).First(temp).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	if temp.Username != "" {
		return false
	}

	return true
}
