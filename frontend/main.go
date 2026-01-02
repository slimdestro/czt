package main

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ccz/handlers"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	port := os.Getenv("FRONTEND_PORT")
	if port == "" {
		port = "8080"
	}
	apiBaseURL := os.Getenv("API_BASE_URL")
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:8081/api"
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			MaxIdleConnsPerHost: 20,
		},
	}

	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	authHandler := &handlers.AuthHandler{
		APIBaseURL: apiBaseURL,
		Tmpl:       tmpl,
		Client:     httpClient,
	}

	profileHandler := &handlers.ProfileHandler{
		APIBaseURL: apiBaseURL,
		Tmpl:       tmpl,
		Client:     httpClient,
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	mux.HandleFunc("/", authHandler.ShowLogin)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			authHandler.ShowLogin(w, r)
		case http.MethodPost:
			authHandler.Login(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			authHandler.ShowSignup(w, r)
		case http.MethodPost:
			authHandler.Signup(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/logout", authHandler.Logout)
	mux.HandleFunc("/auth/google", authHandler.GoogleAuth)
	mux.HandleFunc("/profile", profileHandler.View)
	mux.HandleFunc("/profile/edit", profileHandler.Edit)
	mux.HandleFunc("/profile/save", profileHandler.Save)
	mux.HandleFunc("/profile/cancel", profileHandler.Cancel)
	mux.HandleFunc("/auth/callback", authHandler.AuthCallback)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("frontend server starting", "port", port, "api_url", apiBaseURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("frontend server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down frontend server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("frontend server stopped")
}
