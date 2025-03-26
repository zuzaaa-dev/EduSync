package main

import (
	"EduSync/internal/config"
	"EduSync/internal/delivery/http"
	groupHandler "EduSync/internal/delivery/http/group"
	institutionHandle "EduSync/internal/delivery/http/institution"
	"EduSync/internal/delivery/http/user"
	"EduSync/internal/integration/parser/rksi/group"
	groupRepository "EduSync/internal/repository/group"
	institutionRepository "EduSync/internal/repository/institution"
	userRepository "EduSync/internal/repository/user"
	groupServ "EduSync/internal/service/group"
	institutionServ "EduSync/internal/service/institution"
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
	// TODO: пересмотреть подход реализаций репозитория
	// TODO: Переделать вход с учетом проверки учреждения и проверки группы для студента

	// Инициализируем JWTManager с секретным ключом
	jwtManager := util.NewJWTManager(cfg.JWTSecret)

	// Инициализируем репозитории и сервисы
	userRepo := userRepository.NewUserRepository(db)
	tokenRepo := userRepository.NewTokenRepository(db)
	groupRepo := groupRepository.NewGroupRepository(db)
	institutionRepo := institutionRepository.NewRepository(db)

	groupParser := group.NewGroupParser(cfg.UrlParserRKSI, logger)

	authService := userService.NewAuthService(userRepo, tokenRepo, jwtManager, logger)
	authHandler := user.NewAuthHandler(authService)
	groupService := groupServ.NewGroupService(groupRepo, groupParser, logger)
	//go groupService.StartWorker(100 * time.Second)
	groupHandle := groupHandler.NewGroupHandler(groupService)
	institutionService := institutionServ.NewInstitutionService(institutionRepo, logger)
	institutionHandler := institutionHandle.NewInstitutionHandler(institutionService)
	// Настраиваем маршруты через отдельную функцию в delivery слое
	router := http.SetupRouter(tokenRepo, authHandler, jwtManager, groupHandle, institutionHandler)

	// Запускаем сервер
	port := ":" + cfg.ServerPort
	logger.Infof("🚀 Сервер запущен на порту %s", port)
	log.Fatal(router.Run(port))
}
