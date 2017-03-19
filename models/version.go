package models 

import (
  "github.com/jinzhu/gorm"
  _ "github.com/go-sql-driver/mysql"
)

type Version struct {
  gorm.Model
  Title string `json:"title"`
  Recording string `json:"recording"`
  Notes string `json:"notes"`
  Lyrics string `json:"lyrics"`
  SongID int `json:"song_id"`
  Instruments []Instrument `json:"instruments" gorm:"ForeignKey:VersionID"`
}
