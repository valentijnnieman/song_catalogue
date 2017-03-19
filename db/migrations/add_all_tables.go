package main

import (
  "github.com/jinzhu/gorm"
  _ "github.com/go-sql-driver/mysql"
  "github.com/valentijnnieman/song_catalogue/models"
)

func main() {
  db, err := gorm.Open("mysql", "root@/song_catalogue")
  if err != nil {
    panic("failed to connect database")
  }
  defer db.Close()

  db.AutoMigrate(&models.User{}, &models.Song{}, &models.Version{}, &models.Instrument{})
}
