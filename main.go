package main

import (
  "fmt"
  "os"
  "time"
  "gopkg.in/appleboy/gin-jwt.v2"
  "github.com/gin-gonic/gin"
  "gopkg.in/gin-contrib/cors.v1"
  "github.com/jinzhu/gorm"
  _"github.com/jinzhu/gorm/dialects/postgres"
  "github.com/valentijnnieman/song_catalogue/models"
)

func main() {
  //var song Song
  //var versions []Version
  gin.SetMode(gin.ReleaseMode)

  db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
  fmt.Printf("%s", err)
  defer db.Close()

  r := gin.Default()
	r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"https://valentijnnieman.github.io/song_catalogue_front"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://valentijnnieman.github.io"
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
    AllowOrigins:     []string{"https://valentijnnieman.github.io/song_catalogue_front"},
		AllowMethods:     []string{"PUT", "PATCH", "OPTIONS", "GET", "POST"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
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
  }
  r.Run() // listen and serve on 0.0.0.0:8080
}
