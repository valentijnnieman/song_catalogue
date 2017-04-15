package models

import (
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
)

type Artist struct {
  gorm.Model
  Name string `json:"name"`
  UserID int `json:"UserID"`
  Songs []Song `json:"songs" gorm:"ForeignKey:ArtistID"`
}
