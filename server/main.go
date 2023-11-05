package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.New()

	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://localhost:8080/auth/google/callback"),
	)
	gothic.Store = store

	// API Version 1
	r.GET("/", Index)
	r.GET("/status", GetStatus)

	r.GET("/auth/:provider", BeginAuth)
	r.GET("/auth/:provider/callback", CompleteAuth)

	r.Run(":8080")
}

// GET /
func Index(c *gin.Context) {
	value, err := ReadCookie(c, "user_session")

	if err != nil {
		html := fmt.Sprintf(`<html><body>%v</body></html>`, `<a href="/auth/google">google login</a>`)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Authenticated", "data": value},
	)
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
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Error occurred"},
		)
		return
	}
	SetCookie(c, CreateCookie("user_session"))
	c.JSON(http.StatusOK, gin.H{
		"message": "Authenticated",
		"data":    user},
	)
}

func CreateCookie(name string) http.Cookie {
	return http.Cookie{
		Name:     name,
		Value:    base64.URLEncoding.EncodeToString([]byte(uuid.New().String())),
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func SetCookie(c *gin.Context, cookie http.Cookie) {
	c.SetSameSite(cookie.SameSite)
	c.SetCookie(
		cookie.Name,
		cookie.Value,
		cookie.MaxAge,
		cookie.Path,
		cookie.Domain,
		cookie.Secure,
		cookie.HttpOnly)
}

func ReadCookie(c *gin.Context, name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}

	v, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", err
	}

	return string(v), nil
}
