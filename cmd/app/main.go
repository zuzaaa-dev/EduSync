package main

import (
	"EduSync/internal/config"
	"EduSync/internal/delivery/http"
	chat3 "EduSync/internal/delivery/http/chat"
	email3 "EduSync/internal/delivery/http/email"
	favorite2 "EduSync/internal/delivery/http/favorite"
	groupHandler "EduSync/internal/delivery/http/group"
	institutionHandle "EduSync/internal/delivery/http/institution"
	materialHand "EduSync/internal/delivery/http/material"
	chat4 "EduSync/internal/delivery/http/message"
	schedule2 "EduSync/internal/delivery/http/schedule"
	subjectHandler "EduSync/internal/delivery/http/subject"
	"EduSync/internal/delivery/http/user"
	"EduSync/internal/delivery/ws"
	groupParser "EduSync/internal/integration/parser/rksi/group"
	scheduleParser "EduSync/internal/integration/parser/rksi/schedule"
	teacherParser "EduSync/internal/integration/parser/rksi/teacher"
	"EduSync/internal/repository/chat"
	email2 "EduSync/internal/repository/email"
	favoriteRepository "EduSync/internal/repository/favorite"
	groupRepository "EduSync/internal/repository/group"
	institutionRepository "EduSync/internal/repository/institution"
	materialRepository "EduSync/internal/repository/material"
	scheduleRepository "EduSync/internal/repository/schedule"
	subjectRepository "EduSync/internal/repository/subject"
	userRepository "EduSync/internal/repository/user"
	chat2 "EduSync/internal/service/chat"
	"EduSync/internal/service/email"
	"EduSync/internal/service/favorite"
	groupServ "EduSync/internal/service/group"
	institutionServ "EduSync/internal/service/institution"
	materialServ "EduSync/internal/service/material"
	scheduleServ "EduSync/internal/service/schedule"
	subjectServ "EduSync/internal/service/subject"
	userService "EduSync/internal/service/user"
	"EduSync/internal/util"
	"log"
	"time"
)

// @title          EduSync API
// @version         1.0
// @description     API системы управления образованием

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 "Токен в формате: Bearer {token}"

// @host      localhost:8080
// @BasePath  /api
func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Инициализируем логгер
	util.InitLogger(cfg.LogLevel)
	logger := util.Logger

	logger.Infof(cfg.DatabaseURL)

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
	studentRepo := userRepository.NewStudentRepository(db)
	teacherRepo := userRepository.NewTeacherRepository(db)
	tokenRepo := userRepository.NewTokenRepository(db)
	groupRepo := groupRepository.NewGroupRepository(db)
	subjectRepo := subjectRepository.NewSubjectRepository(db)
	scheduleRepo := scheduleRepository.NewScheduleRepository(db)
	teacherInitionalsRepo := scheduleRepository.NewTeacherInitialsRepository(db)
	institutionRepo := institutionRepository.NewRepository(db)
	emailMaskRepo := institutionRepository.NewEmailMaskRepository(db)
	materialRepo := materialRepository.NewFileRepository(db)
	chatRepo := chat.NewChatRepository(db)
	messageRepo := chat.NewMessageRepository(db)
	favoriteRepo := favoriteRepository.NewFileFavoriteRepository(db)
	pollRepo := chat.NewPollRepository(db)
	emailRepo := email2.NewEmailConfirmationsRepository(db)

	groupParse := groupParser.NewGroupParser(cfg.UrlParserRKSI, logger)
	teacherParse := teacherParser.NewTeacherParser(cfg.UrlParserRKSI, logger)
	scheduleParse := scheduleParser.NewScheduleParser(cfg.UrlParserRKSI, logger)
	subjectService := subjectServ.NewSubjectService(subjectRepo, logger)
	teacherInitionalsService := scheduleServ.NewTeacherInitialsService(teacherInitionalsRepo, logger)
	authService := userService.NewAuthService(userRepo,
		studentRepo,
		teacherRepo,
		tokenRepo,
		emailMaskRepo,
		jwtManager,
		logger,
	)
	materialService := materialServ.NewFileService(materialRepo, messageRepo, chatRepo, logger)
	groupService := groupServ.NewGroupService(groupRepo, groupParse, logger)
	scheduleService := scheduleServ.NewScheduleService(
		scheduleRepo,
		scheduleParse,
		teacherParse,
		subjectService,
		authService,
		groupRepo,
		teacherInitionalsRepo,
		logger,
	)

	hub := ws.NewHub()

	chatSvc := chat2.NewChatService(chatRepo, subjectRepo, userRepo, logger)
	messageSvc := chat2.NewMessageService(messageRepo, logger, hub)
	favoriteSvc := favorite.NewFileFavoriteService(favoriteRepo, materialRepo, messageRepo, chatRepo, logger)
	emailMaskSvc := institutionServ.NewEmailMaskService(emailMaskRepo, logger)
	pollSvc := chat2.NewPollService(pollRepo, chatRepo, logger, hub)
	emailSvc := email.NewSMTPEmailService(
		"mail.nic.ru", 587,
		"postmaster@edusync.ru", "goydaZV1337",
		"no-reply@edusync.ru",
	)
	emailConfirmSVC := email.NewConfirmationService(emailRepo, emailSvc)

	subjectHandle := subjectHandler.NewInstitutionHandler(subjectService)
	authHandler := user.NewAuthHandler(authService)
	groupHandle := groupHandler.NewGroupHandler(groupService)
	go groupService.StartWorker(100 * time.Minute)
	go scheduleService.StartWorkerInitials(100 * time.Minute)
	institutionService := institutionServ.NewInstitutionService(institutionRepo, logger)
	institutionHandler := institutionHandle.NewInstitutionHandler(institutionService, emailMaskSvc)
	scheduleHandler := schedule2.NewScheduleHandler(scheduleService)
	chatHandler := chat3.NewChatHandler(chatSvc)
	messageHandler := chat4.NewMessageHandler(messageSvc)
	materialHandler := materialHand.NewFileHandler(materialService)
	teacherInitionalsHandler := schedule2.NewTeacherInitialsHandler(teacherInitionalsService)
	favoriteHandler := favorite2.NewFileFavoriteHandler(favoriteSvc)
	pollHandler := chat3.NewPollHandler(pollSvc)
	emailHandler := email3.NewConfirmationHandler(emailConfirmSVC)
	// Настраиваем маршруты через отдельную функцию в delivery слое
	router := http.SetupRouter(tokenRepo, chatRepo,
		authHandler,
		jwtManager,
		groupHandle,
		institutionHandler,
		subjectHandle,
		scheduleHandler,
		chatHandler,
		messageHandler,
		materialHandler,
		teacherInitionalsHandler,
		favoriteHandler,
		pollHandler,
		emailHandler,
		logger,
		hub,
	)

	// Запускаем сервер
	port := ":" + cfg.ServerPort
	logger.Infof("🚀 Сервер запущен на порту %s", port)
	log.Fatal(router.Run(port))
}
