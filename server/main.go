package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := httprouter.New()

	// API Version 1
	r.HandlerFunc("GET", "/", Index)
	r.HandlerFunc("GET", "/status", GetStatus)

	r.HandlerFunc("GET", "/auth/:provider", BeginAuth)
	r.HandlerFunc("GET", "/auth/:provider/callback", CompleteAuth)

	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://localhost:8080/auth/google/callback"),
	)

	log.Fatalln(http.ListenAndServe(":8080", r))
}

// GET /
func Index(w http.ResponseWriter, r *http.Request) {
	value, err := ReadCookie(r, "user_session")

	if err != nil {
		html := fmt.Sprintf(`<html><body>%v</body></html>`, `<a href="/auth/google">google login</a>`)
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

// GET /status
func GetStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// GET /auth/:provider
func BeginAuth(w http.ResponseWriter, r *http.Request) {
	fmt.Println(httprouter.ParamsFromContext(r.Context()))
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(gothUser.Email))
		return
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

// GET /auth/:provider/callback
func CompleteAuth(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Error"))
		log.Println(err)
		return
	}
	SetCookie(w, CreateCookie("user_session"))
	w.Write([]byte(user.Email))
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

func SetCookie(w http.ResponseWriter, cookie http.Cookie) {
	http.SetCookie(w, &cookie)
}

func ReadCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	v, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", err
	}

	return string(v), nil
}
