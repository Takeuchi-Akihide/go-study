package main

import (
	"log/slog"
	"net/http"
)

type App struct {
	UserService *UserService
	Logger      *slog.Logger
}

func (app *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.healthHandler)
	mux.HandleFunc("/users", app.usersHandler)
	mux.HandleFunc("/user-summary", app.userSummaryHandler)
	mux.HandleFunc("/jobs", createJobHandler)
	mux.HandleFunc("/jobs/", getJobHandler)
	return mux
}

func (app *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
