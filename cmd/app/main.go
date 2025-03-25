package main

import (
	"EduSync/internal/config"
	"EduSync/internal/delivery/http"
	groupHandler "EduSync/internal/delivery/http/group"
	"EduSync/internal/delivery/http/user"
	"EduSync/internal/integration/parser/rksi/group"
	groupRepository "EduSync/internal/repository/group"
	userRepository "EduSync/internal/repository/user"
	groupServ "EduSync/internal/service/group"
	userService "EduSync/internal/service/user"
	"EduSync/internal/util"
	"log"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Инициализируем логгер
	util.InitLogger(cfg.LogLevel)
	logger := util.Logger

	// Подключаемся к БД
	db, err := config.InitDB(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("Ошибка подключения к БД")
	}
	defer db.Close()

	// Применяем миграции
	config.ApplyMigrations(db, logger)

	// Инициализируем JWTManager с секретным ключом
	jwtManager := util.NewJWTManager(cfg.JWTSecret)

	// Инициализируем репозитории и сервисы
	userRepo := userRepository.NewUserRepository(db)
	tokenRepo := userRepository.NewTokenRepository(db)
	groupRepo := groupRepository.NewGroupRepository(db)
	authService := userService.NewAuthService(userRepo, tokenRepo, jwtManager)
	authHandler := user.NewAuthHandler(authService)
	groupParser := group.NewGroupParser(cfg.UrlParserRKSI)
	groupService := groupServ.NewGroupService(groupRepo, groupParser, 1)
	//go groupService.StartWorker(100 * time.Second)
	groupHandle := groupHandler.NewGroupHandler(groupService)
	// Настраиваем маршруты через отдельную функцию в delivery слое
	router := http.SetupRouter(tokenRepo, authHandler, jwtManager, groupHandle)

	// Запускаем сервер
	port := ":" + cfg.ServerPort
	logger.Infof("🚀 Сервер запущен на порту %s", port)
	log.Fatal(router.Run(port))
}
