package routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ccz/handlers"
	"ccz/middleware"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestProfileRoutes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	h := &handlers.ProfileHandler{DB: db}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/profile", middleware.AuthMiddleware(h.View))
	mux.HandleFunc("/api/profile/save", middleware.AuthMiddleware(h.Save))

	t.Run("ViewProfile_Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
		req.Header.Set("X-User-Email", "test@ex.com")
		w := httptest.NewRecorder()

		rows := sqlmock.NewRows([]string{"full_name", "telephone", "email"}).
			AddRow("Mukul", "123", "test@ex.com")
		mock.ExpectQuery("SELECT full_name, telephone, email FROM users WHERE email=?").
			WithArgs("test@ex.com").WillReturnRows(rows)

		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("ViewProfile_Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("ViewProfile_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
		req.Header.Set("X-User-Email", "none@ex.com")
		w := httptest.NewRecorder()

		mock.ExpectQuery("SELECT full_name, telephone, email FROM users WHERE email=?").
			WithArgs("none@ex.com").WillReturnError(sql.ErrNoRows)

		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("UpdateProfile_Success", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"full_name": "New Name", "telephone": "999"})
		req := httptest.NewRequest(http.MethodPost, "/api/profile/save", bytes.NewBuffer(body))
		req.Header.Set("X-User-Email", "test@ex.com")
		w := httptest.NewRecorder()

		mock.ExpectExec("UPDATE users SET full_name=?, telephone=? WHERE email=?").
			WithArgs("New Name", "999", "test@ex.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("UpdateProfile_MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/profile/save", nil)
		req.Header.Set("X-User-Email", "test@ex.com")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
