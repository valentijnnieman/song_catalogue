package main

import "time"
import "gopkg.in/appleboy/gin-jwt.v2"
import "gopkg.in/gin-contrib/cors.v1"
import "gopkg.in/gin-gonic/gin.v1"

import (
  "github.com/jinzhu/gorm"
  _ "github.com/go-sql-driver/mysql"
  "github.com/valentijnnieman/song_catalogue/models"
)

func main() {
  //var song Song
  //var versions []Version
  //var instruments []Instrument

  db, err := gorm.Open("mysql", "root@/song_catalogue")
  if err != nil {
    panic("failed to connect database")
  }
  defer db.Close()

  r := gin.Default()
	r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:9000"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))

  // the jwt middleware
  authMiddleware := &jwt.GinJWTMiddleware{
      Realm:      "test zone",
      Key:        []byte("secret key"),
      Timeout:    time.Hour,
      MaxRefresh: time.Hour,
      Authenticator: func(userId string, password string, c *gin.Context) (string, bool) {
          if (userId == "admin" && password == "admin") || (userId == "test" && password == "test") {
              return userId, true
          }

          return userId, false
      },
      Authorizator: func(userId string, c *gin.Context) bool {
          if userId == "admin" {
              return true
          }

          return false
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

  r.POST("/login", authMiddleware.LoginHandler)

  auth := r.Group("/auth")
	auth.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:9000"},
		AllowMethods:     []string{"PUT", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
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
    auth.GET("/user/:id", func(c *gin.Context) {
      var user models.User
      db.First(&user)
      db.Model(&user).Related(&user.Songs, "Songs")
      var get_versions []models.Song
      for _, song := range user.Songs {
        db.Model(&song).Related(&song.Versions, "Versions")
        get_versions = append(get_versions, song)
      }
      user.Songs = get_versions
      c.JSON(200, gin.H{
        "user": user,
      })
    })
  }

  //r.GET("/song:id", func(c *gin.Context) {
    //var song models.Song
    //db.First(&song, 1)
    //db.Model(&song).Related(&song.Versions, "Versions")
    //c.JSON(200, gin.H{
      //"song": song,
    //})
  //})

  //r.POST("/song", func(c *gin.Context) {
  //}
  r.Run() // listen and serve on 0.0.0.0:8080
}
