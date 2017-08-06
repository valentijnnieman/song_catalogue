package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Name     string `json:"name"`
	Password string
	Songs    []Song `json:"songs" gorm:"ForeignKey:UserID"`
}
