package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

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
}

type UserStore interface {
	FindAll(ctx context.Context) ([]User, error)
}

func main() {
	dsn := "postgres://dev:password@localhost:5436/app_db?sslmode=disable"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migrate(db); err != nil {
		log.Fatal(err)
	}

	userRepo := &UserRepository{
		DB: db,
	}

	userService := &UserService{
		Store: userRepo,
	}

	app := &App{
		UserService: userService,
	}

	http.HandleFunc("/health", app.healthHandler)
	http.HandleFunc("/users", app.usersHandler)

	log.Println("server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
