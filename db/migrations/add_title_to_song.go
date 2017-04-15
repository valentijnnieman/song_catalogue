package main

import (
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  "github.com/valentijnnieman/song_catalogue/models"
)

func main() {
  db, err := gorm.Open("postgres", "host=localhost user=vaal dbname=song_catalogue sslmode=disable password=testing")
  fmt.Printf("%s", err)
  defer db.Close()
  
  db.AutoMigrate(&models.Song{})
}
