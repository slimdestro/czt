package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"time"
)

type AuthHandler struct {
	APIBaseURL string
	Tmpl       *template.Template
	Client     *http.Client
}

func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	if err := h.Tmpl.ExecuteTemplate(w, "login.html", nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	form := url.Values{}
	form.Add("email", r.FormValue("email"))
	form.Add("password", r.FormValue("password"))

	resp, err := h.Client.PostForm(h.APIBaseURL+"/auth/login", form)
	if err != nil || resp.StatusCode != http.StatusOK {
		if e := h.Tmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Invalid credentials",
		}); e != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	defer resp.Body.Close()

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    result.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *AuthHandler) AuthCallback(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/login?error=unauthorized", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *AuthHandler) ShowSignup(w http.ResponseWriter, r *http.Request) {
	if err := h.Tmpl.ExecuteTemplate(w, "signup.html", nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	form := url.Values{}
	form.Add("email", r.FormValue("email"))
	form.Add("password", r.FormValue("password"))

	resp, err := h.Client.PostForm(h.APIBaseURL+"/auth/signup", form)
	if err != nil || resp.StatusCode != http.StatusCreated {
		if e := h.Tmpl.ExecuteTemplate(w, "signup.html", map[string]string{
			"Error": "Signup failed. Please try again.",
		}); e != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	defer resp.Body.Close()

	http.Redirect(w, r, "/login?signup=success", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, h.APIBaseURL+"/auth/google", http.StatusSeeOther)
}
