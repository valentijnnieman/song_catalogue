package main

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Email    string `json:"email" gorm:"unique"`
	Password string
	Songs    []Song `json:"songs" gorm:"ForeignKey:UserID"`
}

// Struct for binding JSON posts via Gin Gonic bindJSON method
type NEWUSER struct {
	EMAIL    string `json:"email" binding: "required"`
	PASSWORD string `json: "password" binding: "required"`
	SONGS    []Song `json: "songs"`
}
