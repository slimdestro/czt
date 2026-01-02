package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"ccz/middleware"
)

type ProfileHandler struct {
	DB *sql.DB
}

type ProfileResponse struct {
	FullName      string `json:"full_name"`
	Telephone     string `json:"telephone"`
	Email         string `json:"email"`
	EmailDisabled bool   `json:"email_disabled"`
}

func (h *ProfileHandler) View(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || email == "" {
		http.Error(w, "Unauthorized: Identity not found", http.StatusUnauthorized)
		return
	}

	var resp ProfileResponse
	query := "SELECT COALESCE(full_name, ''), COALESCE(telephone, ''), email FROM users WHERE email=?"
	err := h.DB.QueryRowContext(r.Context(), query, email).
		Scan(&resp.FullName, &resp.Telephone, &resp.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User profile not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp.EmailDisabled = true

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ProfileHandler) Save(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		w.Header().Set("Allow", "POST, PUT")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || email == "" {
		http.Error(w, "Unauthorized: Identity not found", http.StatusUnauthorized)
		return
	}

	var input struct {
		FullName  string `json:"full_name"`
		Telephone string `json:"telephone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.FullName == "" {
		http.Error(w, "Full name is required", http.StatusBadRequest)
		return
	}

	query := "UPDATE users SET full_name=?, telephone=? WHERE email=?"
	result, err := h.DB.ExecContext(r.Context(), query, input.FullName, input.Telephone, email)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "No changes made or user not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
