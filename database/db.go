package database

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var Todo *sqlx.DB

type SSLMode string

const SSLModeDisable SSLMode = "disable"

func ConnectAndMigrate(host, port, databaseName, user, password string, sslMode SSLMode) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, databaseName, sslMode)
	DB, err := sqlx.Open("postgres", connStr)

	if err != nil {
		return err
	}

	err = DB.Ping()

	if err != nil {
		return err
	}
	Todo = DB

	return migrateUp(DB)
}

func ShutdownDatabase() error {
	return Todo.Close()
}

func migrateUp(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations/",
		"postgres", driver)

	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
