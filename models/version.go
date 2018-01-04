package models

import (
	"github.com/jinzhu/gorm"
)

type Version struct {
	gorm.Model
	Title     string `json:"title"`
	Recording string `json:"recording"`
	Notes     string `json:"notes"`
	Lyrics    string `json:"lyrics"`
	SongID    int    `json:"song_id"`
}
