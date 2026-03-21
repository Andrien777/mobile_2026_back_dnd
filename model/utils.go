package model

import (
	"dnd_back/api"

	"github.com/jinzhu/gorm"
)

type InternalSpell struct {
	gorm.Model
	api.SpellObject
}

type InternalItem struct {
	gorm.Model
	api.ItemObject
}
