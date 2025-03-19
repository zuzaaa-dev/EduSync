package main

import (
	"EduSync/internal/config"
	"EduSync/internal/util"
	"net/http"
)

func main() {
	// 1️⃣ Загружаем конфигурацию
	cfg := config.LoadConfig()

	// 2️⃣ Инициализируем логирование
	util.InitLogger(cfg.LogLevel)
	logger := util.Logger // Пробрасываем логгер в другие модули

	// 3️⃣ Подключаемся к БД
	db, err := config.InitDB(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("❌ Ошибка подключения к БД")
	}
	defer db.Close()

	// 4️⃣ Применяем миграции
	config.ApplyMigrations(db, logger)

	// 5️⃣ Запускаем сервер
	port := ":" + cfg.ServerPort
	logger.Infof("🚀 Сервер запущен на порту %s", port)
	logger.Fatal(http.ListenAndServe(port, nil)) // В дальнейшем сюда добавится роутинг
}
