package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Email    string `json:"email" gorm:"unique"`
	Password string
	Songs    []Song `json:"songs" gorm:"ForeignKey:UserID"`
}
