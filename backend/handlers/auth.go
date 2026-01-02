package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	DB *sql.DB
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		creds.Email = r.FormValue("email")
		creds.Password = r.FormValue("password")
	}

	if creds.Email == "" || creds.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var id int
	err := h.DB.QueryRowContext(r.Context(), "SELECT id FROM users WHERE email=? AND password=?", creds.Email, creds.Password).Scan(&id)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": creds.Email,
	})
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		creds.Email = r.FormValue("email")
		creds.Password = r.FormValue("password")
	}

	if creds.Email == "" || creds.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	_, err := h.DB.ExecContext(r.Context(), "INSERT INTO users (email, password, provider) VALUES (?, ?, ?)", creds.Email, creds.Password, "local")
	if err != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Google(w http.ResponseWriter, r *http.Request) {
	q := url.Values{}
	q.Set("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	q.Set("redirect_uri", os.Getenv("GOOGLE_REDIRECT_URL"))
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("access_type", "online")
	q.Set("prompt", "select_account")

	http.Redirect(w, r, "https://accounts.google.com/o/oauth2/v2/auth?"+q.Encode(), http.StatusSeeOther)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	frontendURL := os.Getenv("FRONTEND_URL")

	if code == "" {
		http.Redirect(w, r, frontendURL+"/login?error=no_code", http.StatusSeeOther)
		return
	}

	tokenResp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"code":          {code},
		"client_id":     {os.Getenv("GOOGLE_CLIENT_ID")},
		"client_secret": {os.Getenv("GOOGLE_CLIENT_SECRET")},
		"redirect_uri":  {os.Getenv("GOOGLE_REDIRECT_URL")},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		http.Redirect(w, r, frontendURL+"/login?error=token_exchange", http.StatusSeeOther)
		return
	}
	defer tokenResp.Body.Close()

	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&token); err != nil {
		http.Redirect(w, r, frontendURL+"/login?error=json_decode", http.StatusSeeOther)
		return
	}

	req, _ := http.NewRequestWithContext(r.Context(), "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	profileResp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Redirect(w, r, frontendURL+"/login?error=profile_fetch", http.StatusSeeOther)
		return
	}
	defer profileResp.Body.Close()

	var profile struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(profileResp.Body).Decode(&profile); err != nil {
		http.Redirect(w, r, frontendURL+"/login?error=profile_decode", http.StatusSeeOther)
		return
	}

	query := "INSERT INTO users (email, full_name, provider) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE full_name = VALUES(full_name)"
	_, err = h.DB.ExecContext(r.Context(), query, profile.Email, profile.Name, "google")
	if err != nil {
		http.Redirect(w, r, frontendURL+"/login?error=db_error", http.StatusSeeOther)
		return
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": profile.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})
	signedToken, _ := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))

	target := frontendURL + "/auth/callback?token=" + signedToken
	http.Redirect(w, r, target, http.StatusSeeOther)
}
