package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
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

	email := r.FormValue("email")
	pass := r.FormValue("password")

	if email == "" || pass == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var id int
	query := "SELECT id FROM users WHERE email=? AND password=?"
	err := h.DB.QueryRowContext(r.Context(), query, email, pass).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
		"iat":   time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{Token: tokenString})
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	pass := r.FormValue("password")

	if email == "" || pass == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO users(email, password, provider) VALUES(?, ?, ?)"
	_, err := h.DB.ExecContext(r.Context(), query, email, pass, "local")
	if err != nil {
		http.Error(w, "Signup failed", http.StatusBadRequest)
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
