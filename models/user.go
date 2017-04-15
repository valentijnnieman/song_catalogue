package models

import (
  "github.com/jinzhu/gorm"
  _ "github.com/go-sql-driver/mysql"
)

type User struct {
  gorm.Model
  Name string `json:"name"`
  Password string
  Artist Artist `json:"artist" gorm:"ForeignKey:ArtistId"`
}
