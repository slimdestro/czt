package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAuthHandler_Login(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock: %s", err)
	}
	defer db.Close()
	h := &AuthHandler{DB: db}

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		w := httptest.NewRecorder()
		h.Login(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})

	t.Run("Valid Credentials", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "secret")
		form := url.Values{"email": {"test@ex.com"}, "password": {"pass"}}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		mock.ExpectQuery("SELECT id FROM users WHERE email=? AND password=?").
			WithArgs("test@ex.com", "pass").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		h.Login(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestAuthHandler_Signup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock: %s", err)
	}
	defer db.Close()
	h := &AuthHandler{DB: db}

	form := url.Values{"email": {"new@ex.com"}, "password": {"pass"}}
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	mock.ExpectExec("INSERT INTO users").
		WithArgs("new@ex.com", "pass", "local").
		WillReturnResult(sqlmock.NewResult(1, 1))

	h.Signup(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestAuthHandler_GoogleCallback_Errors(t *testing.T) {
	os.Setenv("FRONTEND_URL", "http://frontend.com")
	db, _, _ := sqlmock.New()
	h := &AuthHandler{DB: db}

	t.Run("Missing Code", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/callback", nil)
		w := httptest.NewRecorder()
		h.GoogleCallback(w, req)
		if w.Code != http.StatusSeeOther || !strings.Contains(w.Header().Get("Location"), "error=no_code") {
			t.Errorf("expected redirect with no_code error")
		}
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	h := &AuthHandler{}
	w := httptest.NewRecorder()
	h.Logout(w, httptest.NewRequest(http.MethodGet, "/logout", nil))
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestAuthHandler_Google(t *testing.T) {
	os.Setenv("GOOGLE_CLIENT_ID", "id")
	os.Setenv("GOOGLE_REDIRECT_URL", "http://redirect.com")
	h := &AuthHandler{}
	w := httptest.NewRecorder()
	h.Google(w, httptest.NewRequest(http.MethodGet, "/google", nil))
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}
}
