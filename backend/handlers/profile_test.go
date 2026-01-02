package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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

	t.Run("Unauthorized - No Cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		w := httptest.NewRecorder()
		h.View(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: "test@ex.com"})
		w := httptest.NewRecorder()

		rows := sqlmock.NewRows([]string{"full_name", "telephone", "email"}).
			AddRow("Mukul Kumar", "123456", "test@ex.com")

		mock.ExpectQuery("SELECT (.+) FROM users WHERE email=\\?").
			WithArgs("test@ex.com").
			WillReturnRows(rows)

		h.View(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: "missing@ex.com"})
		w := httptest.NewRecorder()

		mock.ExpectQuery("SELECT (.+) FROM users WHERE email=\\?").
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
		form := url.Values{"full_name": {"Mukul"}, "telephone": {"999"}}
		req := httptest.NewRequest(http.MethodPost, "/profile/save", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session", Value: "test@ex.com"})
		w := httptest.NewRecorder()

		mock.ExpectExec("UPDATE users SET").
			WithArgs("Mukul", "999", "test@ex.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		h.Save(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("DB Failure", func(t *testing.T) {
		form := url.Values{"full_name": {"Mukul"}, "telephone": {"999"}}
		req := httptest.NewRequest(http.MethodPut, "/profile/save", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session", Value: "test@ex.com"})
		w := httptest.NewRecorder()

		mock.ExpectExec("UPDATE users SET").
			WillReturnError(sql.ErrConnDone)

		h.Save(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}
