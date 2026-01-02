package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ccz/middleware"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestProfileHandler_View(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock: %s", err)
	}
	defer db.Close()

	h := &ProfileHandler{DB: db}

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/profile", nil)
		w := httptest.NewRecorder()
		h.View(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})

	t.Run("Unauthorized - No Context Email", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		w := httptest.NewRecorder()
		h.View(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		ctx := context.WithValue(req.Context(), middleware.UserEmailKey, "test@ex.com")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		rows := sqlmock.NewRows([]string{"full_name", "telephone", "email"}).
			AddRow("Mukul Kumar", "123456", "test@ex.com")

		mock.ExpectQuery("SELECT COALESCE\\(full_name, ''\\), COALESCE\\(telephone, ''\\), email FROM users WHERE email=\\?").
			WithArgs("test@ex.com").
			WillReturnRows(rows)

		h.View(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		var resp ProfileResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode profile response: %v", err)
		}
		if resp.Email != "test@ex.com" || resp.FullName != "Mukul Kumar" {
			t.Errorf("unexpected response body")
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		ctx := context.WithValue(req.Context(), middleware.UserEmailKey, "missing@ex.com")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		mock.ExpectQuery("SELECT COALESCE\\(full_name, ''\\), COALESCE\\(telephone, ''\\), email FROM users WHERE email=\\?").
			WithArgs("missing@ex.com").
			WillReturnError(sql.ErrNoRows)

		h.View(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestProfileHandler_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock: %s", err)
	}
	defer db.Close()

	h := &ProfileHandler{DB: db}

	t.Run("Success Update", func(t *testing.T) {
		input := map[string]string{"full_name": "Mukul", "telephone": "999"}
		body, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("failed to marshal input: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/profile/save", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserEmailKey, "test@ex.com")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		mock.ExpectExec("UPDATE users SET full_name=\\?, telephone=\\? WHERE email=\\?").
			WithArgs("Mukul", "999", "test@ex.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		h.Save(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("DB Failure", func(t *testing.T) {
		input := map[string]string{"full_name": "Mukul", "telephone": "999"}
		body, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("failed to marshal input: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/profile/save", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserEmailKey, "test@ex.com")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		mock.ExpectExec("UPDATE users SET full_name=\\?, telephone=\\? WHERE email=\\?").
			WillReturnError(sql.ErrConnDone)

		h.Save(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}
