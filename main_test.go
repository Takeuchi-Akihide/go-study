package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeUserStore struct {
	users []User
	err   error
}

func (s *fakeUserStore) FindAll(ctx context.Context) ([]User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.users, nil
}

func TestUserServiceListUsers(t *testing.T) {
	store := &fakeUserStore{
		users: []User{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
		},
	}

	service := &UserService{
		Store: store,
	}

	users, err := service.ListUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if users[0].Name != "Alice" {
		t.Fatalf("expected Alice, got %s", users[0].Name)
	}
}

func TestUserServiceListUsersError(t *testing.T) {
	store := &fakeUserStore{
		err: errors.New("db error"),
	}

	service := &UserService{
		Store: store,
	}

	_, err := service.ListUsers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUsersHandler(t *testing.T) {
	store := &fakeUserStore{
		users: []User{
			{ID: 1, Name: "Alice"},
		},
	}

	app := &App{
		UserService: &UserService{
			Store: store,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	app.usersHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Alice") {
		t.Fatalf("expected body to contain Alice, got %s", body)
	}
}

func TestUsersHandlerMethodNotAllowed(t *testing.T) {
	app := &App{
		UserService: &UserService{
			Store: &fakeUserStore{},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	rec := httptest.NewRecorder()

	app.usersHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}
