package models

import (
  "github.com/jinzhu/gorm"
)

type User struct {
  gorm.Model
  Name string `json:"name"`
  Password string
  Artist Artist `json:"artist" gorm:"ForeignKey:ArtistId"`
}
