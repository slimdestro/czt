package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"ccz/handlers"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRegisterAuthRoutes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	secret := "testsecret"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	h := &handlers.AuthHandler{DB: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", h.Login)
	mux.HandleFunc("/api/auth/signup", h.Signup)
	mux.HandleFunc("/api/auth/logout", h.Logout)
	mux.HandleFunc("/api/auth/google", h.Google)
	mux.HandleFunc("/api/auth/google/callback", h.GoogleCallback)

	t.Run("Login", func(t *testing.T) {
		mock.ExpectQuery("SELECT id FROM users.*").
			WithArgs("test@ex.com", "pass").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		body, _ := json.Marshal(map[string]string{
			"email":    "test@ex.com",
			"password": "pass",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Signup", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO users.*").
			WithArgs("new@ex.com", "pass", "local").
			WillReturnResult(sqlmock.NewResult(1, 1))

		formData := url.Values{
			"email":    {"new@ex.com"},
			"password": {"pass"},
		}

		req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})

	t.Run("Logout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", w.Code)
		}
	})

	t.Run("Google", func(t *testing.T) {
		os.Setenv("GOOGLE_CLIENT_ID", "id")
		os.Setenv("GOOGLE_REDIRECT_URL", "http://localhost/callback")
		req := httptest.NewRequest(http.MethodGet, "/api/auth/google", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusSeeOther {
			t.Errorf("expected 303, got %d", w.Code)
		}
	})

	t.Run("GoogleCallback_MissingCode", func(t *testing.T) {
		os.Setenv("FRONTEND_URL", "http://localhost")
		req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusSeeOther {
			t.Errorf("expected 303, got %d", w.Code)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
