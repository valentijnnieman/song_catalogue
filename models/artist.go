package models

import (
  "github.com/jinzhu/gorm"
  _ "github.com/go-sql-driver/mysql"
)

type Artist struct {
  gorm.Model
  Name string `json:"name"`
  UserID int `json:"UserID"`
  Songs []Song `json:"songs" gorm:"ForeignKey:ArtistID"`
}
