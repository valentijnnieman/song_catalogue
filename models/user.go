package models

import (
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
)

type User struct {
  gorm.Model
  Name string `json:"name"`
  Password string
  Artist Artist `json:"artist" gorm:"ForeignKey:ArtistId"`
}
