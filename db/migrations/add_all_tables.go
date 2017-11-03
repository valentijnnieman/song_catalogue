package main

import "fmt"
import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/valentijnnieman/song_catalogue/models"
)

func main() {
	var db_url = "host=localhost user=vaal dbname=song_catalogue sslmode=disable password=testing"
	db, err := gorm.Open("postgres", db_url)
	fmt.Printf("%s", err)
	defer db.Close()

	db.AutoMigrate(&models.User{}, &models.Song{}, &models.Version{})
}
