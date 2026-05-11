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
	"strconv"
	"sync"
	"time"

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

type UserSummary struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PostCount         int    `json:"post_count"`
	NotificationCount int    `json:"notification_count"`
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
	mux.HandleFunc("/user-summary", app.userSummaryHandler)

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

func (app *App) userSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idText := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idText)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	summary, err := buildUserSummary(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func buildUserSummary(id int) (UserSummary, error) {
	var wg sync.WaitGroup

	var name string
	var postCount int
	var notificationCount int

	wg.Add(3)

	go func() {
		defer wg.Done()
		name = fetchUserName(id)
	}()

	go func() {
		defer wg.Done()
		postCount = fetchPostCount(id)
	}()

	go func() {
		defer wg.Done()
		notificationCount = fetchNotificationCount(id)
	}()

	wg.Wait()

	return UserSummary{
		ID:                id,
		Name:              name,
		PostCount:         postCount,
		NotificationCount: notificationCount,
	}, nil
}

func fetchUserName(id int) string {
	time.Sleep(500 * time.Millisecond)
	return "Alice"
}

func fetchPostCount(id int) int {
	time.Sleep(500 * time.Millisecond)
	return 12
}

func fetchNotificationCount(id int) int {
	time.Sleep(500 * time.Millisecond)
	return 3
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
