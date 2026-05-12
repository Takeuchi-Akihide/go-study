package main

import "database/sql"

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
