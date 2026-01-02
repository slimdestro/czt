package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRegisterAuthRoutes(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	RegisterAuthRoutes(mux, db)

	paths := []string{
		"/api/auth/login",
		"/api/auth/signup",
		"/api/auth/logout",
		"/api/auth/google",
		"/api/auth/google/callback",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code == http.StatusNotFound {
				t.Errorf("path %s not found", path)
			}
		})
	}
}
