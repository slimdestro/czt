package routes

import (
	"database/sql"
	"net/http"

	"ccz/handlers"
)

func RegisterAuthRoutes(mux *http.ServeMux, db *sql.DB) {
	h := &handlers.AuthHandler{
		DB: db,
	}

	mux.HandleFunc("/api/auth/login", h.Login)
	mux.HandleFunc("/api/auth/signup", h.Signup)
	mux.HandleFunc("/api/auth/logout", h.Logout)
	mux.HandleFunc("/api/auth/google", h.Google)
	mux.HandleFunc("/api/auth/google/callback", h.GoogleCallback)
}
