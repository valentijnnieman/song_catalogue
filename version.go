package main

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

// Struct for binding JSON posts via Gin Gonic bindJSON method
type NEWVERSION struct {
	TITLE     string `json:"title" binding:"required"`
	RECORDING string `json:"recording" binding:"required"`
	NOTES     string `json:"notes" binding:"required"`
	LYRICS    string `json:"lyrics" binding:"required"`
	SONG_ID   int    `json:"song_id" binding:"required"`
}
