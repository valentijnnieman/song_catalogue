package models

import (
	"github.com/jinzhu/gorm"
)

type Song struct {
	gorm.Model
	Title    string    `json:"title"`
	UserID   int       `json:"user_id"`
	Versions []Version `json:"versions" gorm:"ForeignKey:SongID"`
}
