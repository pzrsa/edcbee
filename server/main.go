package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

var APIv1 *gin.RouterGroup

func main() {
	router := gin.New()

	APIv1 = router.Group("api/v1")

	// API Version 1
	GetStatus(APIv1)
	Login(APIv1)

	router.Run(":8080")
}

func Login(router *gin.RouterGroup) {
	router.POST("/login", func(c *gin.Context) {
		// from https://www.alexedwards.net/blog/basic-authentication-in-go
		if username, password, ok := c.Request.BasicAuth(); ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte("p"))
			expectedPasswordHash := sha256.Sum256([]byte("m"))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				c.JSON(http.StatusOK, gin.H{"message": "Authenticated"})
				return
			}
		}
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
	})
}

func GetStatus(router *gin.RouterGroup) {
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
}
