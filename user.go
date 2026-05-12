package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

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

type UserStore interface {
	FindAll(ctx context.Context) ([]User, error)
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
