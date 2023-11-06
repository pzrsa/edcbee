package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type H map[string]interface{}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := mux.NewRouter()

	// API Version 1
	r.HandleFunc("/", Index).Methods("GET")
	r.HandleFunc("/status", GetStatus).Methods("GET")

	r.HandleFunc("/auth/{provider}", BeginAuth).Methods("GET")
	r.HandleFunc("/auth/{provider}/callback", CompleteAuth).Methods("GET")

	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://localhost:8080/auth/google/callback"),
	)
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println(fmt.Sprintf("listening on %s...", srv.Addr))
	log.Fatalln(srv.ListenAndServe())
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(H{"message": "Authenticated", "data": value})
}

// GET /status
func GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(H{"message": "yo"})
}

// GET /auth/{provider}
func BeginAuth(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

// GET /auth/{provider}/callback
func CompleteAuth(w http.ResponseWriter, r *http.Request) {
	_, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		log.Println("Error:", err)
		json.NewEncoder(w).Encode(H{"message": "Error"})
		return
	}
	SetCookie(w, CreateCookie("user_session"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
