package model

import "github.com/jinzhu/gorm"

type InternalCharacterShare struct {
	gorm.Model
	CharacterID        uint   `gorm:"not null;index"`
	SharedWithUsername string `gorm:"not null;index"`
}
