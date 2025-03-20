package config

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
)

// ApplyMigrations автоматически применяет миграции при запуске
func ApplyMigrations(db *sql.DB, logger *logrus.Logger) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Fatalf("Ошибка создания драйвера миграции: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		logger.Fatalf("Ошибка создания экземпляра миграции: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Fatalf("Ошибка применения миграций: %v", err)
	}

	logger.Info("Миграции успешно применены")
}
