package main

import (
	"fmt"

	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"golang.org/x/crypto/bcrypt"

	"github.com/appleboy/gin-jwt"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func HashString(s string) string {
	hs, error := bcrypt.GenerateFromPassword([]byte(s), 14)
	if error != nil {
		fmt.Errorf("%s", error)
	}
	return string(hs[:])
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

	db.AutoMigrate(&User{}, &Song{}, &Version{})

	// Create a demo/debug account on server start
	// var user = models.User{}
	// db.FirstOrCreate(&user, models.User{Email: "demodemodemo", Password: HashString("demo123")})

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String("us-east-2")},
		SharedConfigState: session.SharedConfigEnable,
	}))
	uploader := s3manager.NewUploader(sess)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			// return origin == allowed_origin
			return true
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
			var user User
			db.Where("email = ?", userId).First(&user)
			pwdError := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if pwdError == nil {
				return userId, true
			}
			return userId, false
		},
		Authorizator: func(email string, c *gin.Context) bool {
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"message": message,
			})
		},
		TokenLookup: "header:Authorization",
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	r.POST("/login", authMiddleware.LoginHandler)

	r.POST("/register", func(c *gin.Context) {
		var newUser NEWUSER
		c.BindJSON(&newUser)

		if len(newUser.EMAIL) < 6 || len(newUser.PASSWORD) < 6 {
			c.JSON(401, gin.H{
				"message": "Email or password invalid",
			})
		} else {
			var user = User{Email: newUser.EMAIL, Password: HashString(newUser.PASSWORD)}
			if err := db.Create(&user).Error; err != nil {
				c.JSON(409, gin.H{
					"message": err,
				})
			} else {
				c.JSON(200, gin.H{
					"message": "Successfully created new account",
				})
			}
		}
	})

	auth := r.Group("/auth")
	auth.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/refresh_token", authMiddleware.RefreshHandler)

		auth.POST("/reset-password", func(c *gin.Context) {
			email := c.PostForm("email")
			oldPassword := c.PostForm("password")
			newPassword := c.PostForm("newPassword")

			claims := jwt.ExtractClaims(c)

			if email != claims["id"] {
				fmt.Println("email doesn't match!")
				c.JSON(401, gin.H{
					"message": "Email or password is incorrect",
				})
			} else {
				var user User
				db.Where("email = ?", claims["id"]).First(&user)

				pwdError := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
				if pwdError != nil {
					c.JSON(401, gin.H{
						"message": "Email or password is incorrect",
					})
				} else {
					user.Password = HashString(newPassword)
					db.Save(&user)
					c.JSON(200, gin.H{
						"message": "Successfully reset password!",
					})
				}
			}

		})

		auth.GET("/songs", func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			var user User

			db.Where("email = ?", claims["id"]).First(&user)
			db.Model(&user).Related(&user.Songs, "Songs")
			var getVersions []Song
			for _, song := range user.Songs {
				db.Model(&song).Related(&song.Versions, "Versions")
				getVersions = append(getVersions, song)
			}
			user.Songs = getVersions
			c.JSON(200, gin.H{
				"songs":   user.Songs,
				"message": "Successfully got songs!",
			})
		})

		auth.GET("/songs/:song_id", func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			var user User

			db.Where("email = ? id = ?", claims["id"], c.Param("song_id")).First(&user)
			db.Model(&user).Related(&user.Songs, "Songs")
			var getVersions []Song
			for _, song := range user.Songs {
				db.Model(&song).Related(&song.Versions, "Versions")
				getVersions = append(getVersions, song)
			}
			user.Songs = getVersions
			c.JSON(200, gin.H{
				"songs":   user.Songs,
				"message": "Successfully got songs!",
			})
		})

		// TODO: auth.PATCH("/songs/:song_id")

		auth.POST("/songs/create", func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			var user User
			db.Where("email = ?", claims["id"]).First(&user)

			var new_song NEWSONG
			c.BindJSON(&new_song)

			var song = Song{Title: new_song.TITLE, UserID: int(user.ID), Versions: new_song.VERSIONS}
			if err := db.Create(&song).Error; err != nil {
				c.JSON(409, gin.H{
					"message": err,
				})
			} else {
				c.JSON(200, gin.H{
					"song":    song,
					"message": "Successfully created new song!",
				})
			}
		})
		auth.DELETE("/songs/:song_id/delete", func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			var user User
			db.Where("email = ?", claims["id"]).First(&user)

			var song Song

			if err := db.Where("user_id = ?", user.ID).First(&song, c.Param("song_id")).Error; err != nil {
				c.JSON(409, gin.H{
					"message": err,
				})
			} else {
				db.Delete(&song)
				c.JSON(200, gin.H{
					"message": "Successfully deleted song!",
				})
			}
		})

		// TODO: GET versions && versions/:version_id (if/when needed)

		auth.POST("/versions/create", func(c *gin.Context) {
			var newVersion NEWVERSION
			c.BindJSON(&newVersion)

			var version = Version{Title: newVersion.TITLE, Recording: newVersion.RECORDING, Notes: newVersion.NOTES, Lyrics: newVersion.LYRICS, SongID: newVersion.SONG_ID}

			if err := db.Create(&version).Error; err != nil {
				c.JSON(409, gin.H{
					"message": err,
				})
			} else {
				c.JSON(200, gin.H{
					"version": version,
					"message": "Successfully created new version!",
				})
			}
		})

		auth.PATCH("/versions/:version_id/update", func(c *gin.Context) {
			var newVersion NEWVERSION
			c.BindJSON(&newVersion)

			var current_version Version
			db.First(&current_version, c.Param("version_id"))
			if err := db.Model(&current_version).Updates(newVersion).Error; err != nil {
				c.JSON(409, gin.H{
					"message": err,
				})
			} else {
				c.JSON(200, gin.H{
					"version": current_version,
					"message": "Successfully updated version!",
				})
			}
		})

		auth.DELETE("/versions/:version_id/delete", func(c *gin.Context) {
			song_id := c.PostForm("song_id")
			var version Version

			db.Where("song_id = ?", song_id).First(&version, c.Param("version_id"))

			if err := db.Delete(&version).Error; err != nil {
				c.JSON(409, gin.H{
					"message": err,
				})
			} else {
				c.JSON(200, gin.H{
					"message": "Successfully deleted version!",
				})
			}
		})

		// RPC style endpoint for easier file uploading
		// Gin Gonic router complains if route has version_id as param, so we take it from post body
		auth.POST("/versions/recording", func(c *gin.Context) {
			version_id := c.PostForm("version_id")
			song_id := c.PostForm("song_id")
			file, _ := c.FormFile("file")
			openfile, _ := file.Open()
			new_filepath := "recording_" + song_id + version_id + "_" + file.Filename

			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String("song-catalogue-storage"),
				Key:    aws.String(new_filepath),
				Body:   openfile,
			})

			if err != nil {
				fmt.Errorf("%s", err)
				c.JSON(500, gin.H{
					"message": err,
				})
			} else {
				songpath := "https://s3.us-east-2.amazonaws.com/song-catalogue-storage/" + new_filepath

				var version Version
				db.Where("song_id = ?", song_id).First(&version, version_id)

				version.Recording = songpath

				db.Save(&version)

				c.JSON(200, gin.H{
					"version": version,
					"message": "Successfully uploaded new recording!",
				})
			}
		})
	}
	r.Run()
}
