package main

import (
  "fmt"
  //"os"
  "time"
  "gopkg.in/appleboy/gin-jwt.v2"
  "github.com/gin-gonic/gin"
  "gopkg.in/gin-contrib/cors.v1"
  "github.com/jinzhu/gorm"
  _"github.com/jinzhu/gorm/dialects/postgres"
  "github.com/valentijnnieman/song_catalogue/models"
)

func main() {
  var db_url string;
  if gin.Mode() == "debug" {
    db_url = "host=localhost user=valentijnnieman dbname=song_catalogue sslmode=disable password=testing"
  } else {
    db_url = os.Getenv("DATABASE_URL")
  }
  db, err := gorm.Open("postgres", db_url)
  fmt.Printf("%s", err)
  defer db.Close()

  r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST"},
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
        if (user.Password == password) {
          return userId, true
        }
        return userId, false
      },
      Authorizator: func(userId string, c *gin.Context) bool {
        return true
          //if userId == "Hans" {
              //return true
          //}

          //return false
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
		AllowMethods:     []string{"PUT", "PATCH", "OPTIONS", "GET", "POST"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))

  auth.Use(authMiddleware.MiddlewareFunc())
  {
    auth.GET("/refresh_token", authMiddleware.RefreshHandler)

    auth.GET("/test", func(c *gin.Context) {
      c.JSON(200, gin.H{
        "test": "yay",
      })
    })
    auth.GET("/artist/:id", func(c *gin.Context) {
      var artist models.Artist
      db.First(&artist)
      db.Model(&artist).Related(&artist.Songs, "Songs")
      var get_versions []models.Song
      for _, song := range artist.Songs {
        db.Model(&song).Related(&song.Versions, "Versions")
        get_versions = append(get_versions, song)
      }
      artist.Songs = get_versions
      c.JSON(200, gin.H{
        "artist": artist,
      })
    })
    auth.POST("/artist/:id/song/:id", func(c *gin.Context) {

    })
    auth.PUT("/artist/:id/song/:id", func(c *gin.Context) {

    })
    auth.DELETE("/artist/:id/song/:id", func(c *gin.Context) {

    })
  }
  r.Run() // listen and serve on 0.0.0.0:8080
}
