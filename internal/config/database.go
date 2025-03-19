package config

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// InitDB подключается к базе данных
func InitDB(databaseURL string, logger *logrus.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Fatalf("❌ Ошибка подключения к БД: %v", err)
		return nil, err
	}

	// Ждем, пока БД будет доступна
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err != nil {
			logger.Warnf("⏳ Ожидание подключения к БД (%d/10)...", i+1)
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	logger.Info("✅ Подключение к БД успешно установлено")
	return db, nil
}
