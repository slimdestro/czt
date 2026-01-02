package routes

import (
	"database/sql"
	"net/http"

	"ccz/handlers"
	"ccz/middleware"
)

func RegisterProfileRoutes(mux *http.ServeMux, db *sql.DB) {
	h := &handlers.ProfileHandler{
		DB: db,
	}

	mux.HandleFunc("/api/profile", middleware.AuthMiddleware(h.View))
	mux.HandleFunc("/api/profile/save", middleware.AuthMiddleware(h.Save))
}
