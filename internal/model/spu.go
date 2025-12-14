package model

import (
	"gorm.io/gorm"
)

// SPU (Standard Product Unit) represents a generic product concept without specific attributes.
type SPU struct {
	gorm.Model
	Name        string `gorm:"not null;type:varchar(255)" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	CategoryID  uint   `gorm:"index" json:"category_id"` // Foreign key to a (yet to be defined) Category model
}
