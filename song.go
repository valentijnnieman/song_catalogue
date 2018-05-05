package main

import (
	"github.com/jinzhu/gorm"
)

type Song struct {
	gorm.Model
	Title    string    `json:"title"`
	UserID   int       `json:"user_id"`
	Versions []Version `json:"versions" gorm:"ForeignKey:SongID"`
}

// Struct for binding JSON posts via Gin Gonic bindJSON method
type NEWSONG struct {
	TITLE    string    `json:"title" binding:"required"`
	VERSIONS []Version `json:"versions"`
}
