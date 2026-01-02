package handlers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
)

type ProfileHandler struct {
	APIBaseURL string
	Tmpl       *template.Template
	Client     *http.Client
}

type ProfileViewModel struct {
	FullName      string `json:"full_name"`
	Telephone     string `json:"telephone"`
	Email         string `json:"email"`
	EmailDisabled bool   `json:"email_disabled"`
}

func (h *ProfileHandler) getProfile(r *http.Request) (*ProfileViewModel, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return nil, false
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, h.APIBaseURL+"/profile", nil)
	if err != nil {
		return nil, false
	}

	req.Header.Set("Authorization", "Bearer "+cookie.Value)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false
	}

	var vm ProfileViewModel
	if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
		return nil, false
	}
	return &vm, true
}

func (h *ProfileHandler) View(w http.ResponseWriter, r *http.Request) {
	vm, ok := h.getProfile(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h.Tmpl.ExecuteTemplate(w, "profile_view.html", vm)
}

func (h *ProfileHandler) Edit(w http.ResponseWriter, r *http.Request) {
	vm, ok := h.getProfile(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h.Tmpl.ExecuteTemplate(w, "profile_edit.html", vm)
}

func (h *ProfileHandler) Save(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	payload := map[string]string{
		"full_name": r.FormValue("full_name"),
		"telephone": r.FormValue("telephone"),
	}

	reqBody, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.APIBaseURL+"/api/profile/save", bytes.NewReader(reqBody))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+cookie.Value)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.Client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Redirect(w, r, "/profile/edit?error=update_failed", http.StatusSeeOther)
		return
	}
	defer resp.Body.Close()

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *ProfileHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
