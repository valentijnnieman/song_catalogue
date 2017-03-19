package models

import (
  "github.com/jinzhu/gorm"
  _ "github.com/go-sql-driver/mysql"
)

type User struct {
  gorm.Model
  Name string `json:"name"`
  Songs []Song `json:"songs" gorm:"ForeignKey:UserID"`
}
