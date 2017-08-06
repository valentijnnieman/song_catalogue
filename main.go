package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/valentijnnieman/song_catalogue/models"
	"gopkg.in/appleboy/gin-jwt.v2"
	"gopkg.in/gin-contrib/cors.v1"
	"os"
	"time"
)

type NEWSONG struct {
	TITLE    string           `json:"title" binding:"required"`
	VERSIONS []models.Version `json:"versions"`
}
type NEWVERSION struct {
	TITLE     string `json:"title" binding:"required"`
	RECORDING string `json:"recording" binding:"required"`
	NOTES     string `json:"notes" binding:"required"`
	LYRICS    string `json:"lyrics" binding:"required"`
	SONG_ID   int    `json:"song_id" binding:"required"`
}

func main() {
	var db_url string
	if gin.Mode() == "debug" {
		db_url = "host=localhost user=vaal dbname=song_catalogue sslmode=disable password=testing"
	} else {
		db_url = os.Getenv("DATABASE_URL")
	}
	db, err := gorm.Open("postgres", db_url)
	fmt.Printf("%s", err)
	defer db.Close()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if gin.Mode() == "debug" {
				return origin == "http://localhost:9000"
			} else {
				return origin == "https://valentijnnieman.github.io"
			}
		},
		MaxAge: 12 * time.Hour,
	}))

	// the jwt middleware
	authMiddleware := &jwt.GinJWTMiddleware{
		Realm:      "song_catalogue",
		Key:        []byte("secret key"),
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		Authenticator: func(userId string, password string, c *gin.Context) (string, bool) {
			var user models.User
			db.Where("name = ?", userId).First(&user)
			if user.Password == password {
				return userId, true
			}
			return userId, false
		},
		Authorizator: func(userId string, c *gin.Context) bool {
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup: "header:Authorization",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",
	}
	// USED FOR TESTING & DEMO

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"ping": "success!",
		})
	})

	r.POST("/login", authMiddleware.LoginHandler)

	auth := r.Group("/auth")
	auth.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"PUT", "PATCH", "OPTIONS", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/refresh_token", authMiddleware.RefreshHandler)

		auth.GET("/songs", func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			var user models.User

			db.Where("name = ?", claims["id"]).First(&user)
			db.Model(&user).Related(&user.Songs, "Songs")
			var get_versions []models.Song
			for _, song := range user.Songs {
				db.Model(&song).Related(&song.Versions, "Versions")
				get_versions = append(get_versions, song)
			}
			user.Songs = get_versions
			c.JSON(200, gin.H{
				"songs": user.Songs,
			})
		})

		auth.POST("/song/create", func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			var user models.User
			db.Where("name = ?", claims["id"]).First(&user)

			var new_song NEWSONG
			c.BindJSON(&new_song)

			var song = models.Song{Title: new_song.TITLE, UserID: int(user.ID), Versions: new_song.VERSIONS}
			db.Create(&song)

			c.JSON(200, gin.H{
				"song": song,
			})
		})
		auth.DELETE("/song/:song_id/delete", func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			var user models.User
			db.Where("name = ?", claims["id"]).First(&user)

			var song models.Song

			db.Where("user_id = ?", user.ID).First(&song, c.Param("song_id"))

			db.Delete(&song)
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})
		auth.POST("/version/create", func(c *gin.Context) {
			var new_version NEWVERSION
			c.BindJSON(&new_version)

			fmt.Printf("%s \n", &new_version)

			var version = models.Version{Title: new_version.TITLE, Recording: new_version.RECORDING, Notes: new_version.NOTES, Lyrics: new_version.LYRICS, SongID: new_version.SONG_ID}

			db.Create(&version)
			c.JSON(200, gin.H{
				"version": version,
			})
		})
		auth.PATCH("/version/:version_id/update", func(c *gin.Context) {
			var new_version NEWVERSION
			c.BindJSON(&new_version)

			var current_version models.Version
			db.First(&current_version, c.Param("version_id"))
			db.Model(&current_version).Updates(new_version)

			c.JSON(200, gin.H{
				"version": current_version,
			})
		})
		auth.DELETE("/song/:song_id/version/:version_id/delete", func(c *gin.Context) {
			var version models.Version

			db.Where("song_id = ?", c.Param("song_id")).First(&version, c.Param("version_id"))

			db.Delete(&version)
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})
	}
	r.Run() // listen and serve on 0.0.0.0:8080
}
