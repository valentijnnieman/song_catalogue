package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"golang.org/x/crypto/bcrypt"
	//"github.com/aws/aws-sdk-go/service/s3"
	"os"
	"time"

	"github.com/appleboy/gin-jwt"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/valentijnnieman/song_catalogue/api/models"
)

type NEWUSER struct {
	EMAIL    string        `json:"email" binding: "required"`
	PASSWORD string        `json: "password" binding: "required"`
	SONGS    []models.Song `json: "songs"`
}

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

func HashString(s string) string {
	hs, error := bcrypt.GenerateFromPassword([]byte(s), 14)
	if error != nil {
		fmt.Errorf("%s", error)
	}
	return string(hs[:])
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
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

	db.AutoMigrate(&models.User{}, &models.Song{}, &models.Version{})

	// Create a demo account on server start
	var user = models.User{}
	db.FirstOrCreate(&user, models.User{Email: "demodemodemo", Password: HashString("demo123")})

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String("us-east-2")},
		SharedConfigState: session.SharedConfigEnable,
	}))
	uploader := s3manager.NewUploader(sess)
	//downloader := s3manager.NewDownloader(sess)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:3000"
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
				"code":    code,
				"message": message,
			})
		},
		TokenLookup: "header:Authorization",
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"ping": "success!",
		})
	})

	r.POST("/login", authMiddleware.LoginHandler)

	r.POST("/register", func(c *gin.Context) {
		var newUser NEWUSER
		c.BindJSON(&newUser)
		fmt.Println(newUser)

		var user = models.User{Email: newUser.EMAIL, Password: HashString(newUser.PASSWORD)}
		db.Create(&user)
		c.JSON(200, gin.H{
			"success": true,
		})
	})

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

			fmt.Println("claims[id]: ", claims["id"])
			db.Where("email = ?", claims["id"]).First(&user)
			db.Model(&user).Related(&user.Songs, "Songs")
			var getVersions []models.Song
			for _, song := range user.Songs {
				db.Model(&song).Related(&song.Versions, "Versions")
				getVersions = append(getVersions, song)
			}
			user.Songs = getVersions
			c.JSON(200, gin.H{
				"songs": user.Songs,
			})
		})

		auth.POST("/song/create", func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			var user models.User
			fmt.Println("claims[id]: ", claims["id"])
			db.Where("email = ?", claims["id"]).First(&user)

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
			db.Where("email = ?", claims["id"]).First(&user)

			var song models.Song

			db.Where("user_id = ?", user.ID).First(&song, c.Param("song_id"))

			db.Delete(&song)
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})

		auth.POST("/version/recording", func(c *gin.Context) {
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
				// Print the error and exit.
				exitErrorf("Unable to upload! %s \n", err)
			}

			songpath := "https://s3.us-east-2.amazonaws.com/song-catalogue-storage/" + new_filepath

			var version models.Version
			db.Where("song_id = ?", song_id).First(&version, version_id)

			version.Recording = songpath

			db.Save(&version)

			fmt.Printf("Successfully uploaded! ")
			c.JSON(200, gin.H{
				"version": version,
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
