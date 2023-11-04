package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.New()

	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://localhost:8080/auth/google/callback"),
	)

	// API Version 1
	r.GET("/", Index)
	r.GET("/status", GetStatus)

	r.GET("/auth/:provider", BeginAuth)
	r.GET("/auth/:provider/callback", CompleteAuth)

	r.Run(":8080")
}

// GET /
func Index(c *gin.Context) {
	html := fmt.Sprintf(`<html><body>%v</body></html>`, `<a href="/v1/auth/google">google login</a>`)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// GET /status
func GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

// GET /auth/:provider
func BeginAuth(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", c.Param("provider"))
	c.Request.URL.RawQuery = q.Encode()
	if gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Authenticated", "data": gothUser},
		)
		return
	} else {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}
}

// GET /auth/:provider/callback
func CompleteAuth(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", c.Param("provider"))
	c.Request.URL.RawQuery = q.Encode()
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		log.Println(c.Writer, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Authenticated", "data": user},
	)
}
