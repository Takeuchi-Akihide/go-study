package main

import (
	"database/sql"
	"encoding/json"
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
	DB *sql.DB
}

func main() {
	dsn := "postgres://dev:password@localhost:5436/app_db?sslmode=disable"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	app := &App{
		DB: db,
	}

	http.HandleFunc("/health", app.healthHandler)
	http.HandleFunc("/users", app.usersHandler)

	log.Println("server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (app *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (app *App) usersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	users, err := app.findUsers()
	if err != nil {
		log.Println("failed to find users:", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (app *App) findUsers() ([]User, error) {
	rows, err := app.DB.Query(`
    SELECT id, name
    FROM users
    ORDER BY id
  `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name)
		if err != nil {
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
