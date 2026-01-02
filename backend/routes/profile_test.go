package routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestProfileRoutes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	RegisterProfileRoutes(mux, db)

	t.Run("ViewProfile_Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
		req.Header.Set("X-User-Email", "test@ex.com")
		w := httptest.NewRecorder()

		rows := sqlmock.NewRows([]string{"full_name", "telephone", "email"}).
			AddRow("Mukul", "123", "test@ex.com")
		mock.ExpectQuery("SELECT (.+) FROM users").WithArgs("test@ex.com").WillReturnRows(rows)

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
			t.Errorf("expected 401")
		}
	})

	t.Run("ViewProfile_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
		req.Header.Set("X-User-Email", "none@ex.com")
		w := httptest.NewRecorder()
		mock.ExpectQuery("SELECT (.+) FROM users").WillReturnError(sql.ErrNoRows)

		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404")
		}
	})

	t.Run("UpdateProfile_Success", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"full_name": "New Name", "telephone": "999"})
		req := httptest.NewRequest(http.MethodPost, "/api/profile/update", bytes.NewBuffer(body))
		req.Header.Set("X-User-Email", "test@ex.com")
		w := httptest.NewRecorder()

		mock.ExpectExec("UPDATE users SET").WithArgs("New Name", "999", "test@ex.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200")
		}
	})

	t.Run("UpdateProfile_MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/profile/update", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405")
		}
	})
}
