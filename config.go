package main

import "os"

type Config struct {
	Port string
	DSN  string
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
