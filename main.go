package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

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

	startWorkers(3)

	addr := ":" + cfg.Port

	logger.Info("server started", "addr", addr)
	if err := http.ListenAndServe(addr, app.routes()); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
