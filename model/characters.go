package model

import (
	"dnd_back/api"

	"github.com/jinzhu/gorm"
)

type InternalCharacter struct {
	gorm.Model
	api.CharacterObject
	Owner string
}
