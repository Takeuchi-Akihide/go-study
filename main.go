package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	Port string
	DSN  string
}

type HealthResponse struct {
	Status string `json:"status"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type App struct {
	UserService *UserService
	Logger      *slog.Logger
}

type UserStore interface {
	FindAll(ctx context.Context) ([]User, error)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := loadConfig()

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		logger.Error("failed to open db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping db", "error", err)
		os.Exit(1)
	}

	if err := migrate(db); err != nil {
		logger.Error("failed to migrate db", "error", err)
		os.Exit(1)
	}

	userRepo := &UserRepository{
		DB: db,
	}

	userService := &UserService{
		Store: userRepo,
	}

	app := &App{
		UserService: userService,
		Logger:      logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.healthHandler)
	mux.HandleFunc("/users", app.usersHandler)

	addr := ":" + cfg.Port

	logger.Info("server started", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://dev:password@localhost:5436/app_db?sslmode=disable"
	}

	return Config{
		Port: port,
		DSN:  dsn,
	}
}

func migrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL
	);
	
	INSERT INTO users (name)
	SELECT 'Alice'
	WHERE NOT EXISTS (SELECT 1 FROM users WHERE name = 'Alice');
	
	INSERT INTO users (name)
	SELECT 'Bob'
	WHERE NOT EXISTS (SELECT 1 FROM users WHERE name = 'Bob');
	`

	_, err := db.Exec(query)
	return err
}

func (app *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (app *App) usersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	users, err := app.UserService.ListUsers(r.Context())
	if err != nil {
		log.Println("failed to find users:", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, users)
}

type UserService struct {
	Store UserStore
}

func (s *UserService) ListUsers(ctx context.Context) ([]User, error) {
	users, err := s.Store.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) FindAll(ctx context.Context) ([]User, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("failed to encode json:", err)
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}

var ErrUserNotFound = errors.New("user not found")
