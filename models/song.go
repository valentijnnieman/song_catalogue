package models 

import (
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
)

type Song struct {
  gorm.Model
  Title string `json:"title"`
  ArtistID int `json:"artist_id"`
  Versions []Version `json:"versions" gorm:"ForeignKey:SongID"`
}
