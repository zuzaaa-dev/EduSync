package main

import (
	"EduSync/internal/config"
	"EduSync/internal/delivery/http"
	groupHandler "EduSync/internal/delivery/http/group"
	institutionHandle "EduSync/internal/delivery/http/institution"
	schedule2 "EduSync/internal/delivery/http/schedule"
	subjectHandler "EduSync/internal/delivery/http/subject"
	"EduSync/internal/delivery/http/user"
	groupParser "EduSync/internal/integration/parser/rksi/group"
	scheduleParser "EduSync/internal/integration/parser/rksi/schedule"
	groupRepository "EduSync/internal/repository/group"
	institutionRepository "EduSync/internal/repository/institution"
	scheduleRepository "EduSync/internal/repository/schedule"
	subjectRepository "EduSync/internal/repository/subject"
	userRepository "EduSync/internal/repository/user"
	groupServ "EduSync/internal/service/group"
	institutionServ "EduSync/internal/service/institution"
	scheduleServ "EduSync/internal/service/schedule"
	subjectServ "EduSync/internal/service/subject"
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
	studentRepo := userRepository.NewStudentRepository(db)
	teacherRepo := userRepository.NewTeacherRepository(db)
	tokenRepo := userRepository.NewTokenRepository(db)
	groupRepo := groupRepository.NewGroupRepository(db)
	subjectRepo := subjectRepository.NewSubjectRepository(db)
	scheduleRepo := scheduleRepository.NewScheduleRepository(db)
	institutionRepo := institutionRepository.NewRepository(db)
	emailMaskRepo := institutionRepository.NewEmailMaskRepository(db)

	groupParse := groupParser.NewGroupParser(cfg.UrlParserRKSI, logger)
	scheduleParse := scheduleParser.NewScheduleParser(cfg.UrlParserRKSI, logger)
	subjectService := subjectServ.NewSubjectService(subjectRepo, logger)
	authService := userService.NewAuthService(userRepo,
		studentRepo,
		teacherRepo,
		tokenRepo,
		emailMaskRepo,
		jwtManager,
		logger,
	)
	groupService := groupServ.NewGroupService(groupRepo, groupParse, logger)
	scheduleService := scheduleServ.NewScheduleService(
		scheduleRepo,
		scheduleParse,
		subjectService,
		authService,
		groupRepo,
		logger,
	)
	emailMaskSvc := institutionServ.NewEmailMaskService(emailMaskRepo, logger)
	subjectHandle := subjectHandler.NewInstitutionHandler(subjectService)
	authHandler := user.NewAuthHandler(authService)
	groupHandle := groupHandler.NewGroupHandler(groupService)
	//go groupService.StartWorker(100 * time.Second)
	institutionService := institutionServ.NewInstitutionService(institutionRepo, logger)
	institutionHandler := institutionHandle.NewInstitutionHandler(institutionService, emailMaskSvc)
	scheduleHandler := schedule2.NewScheduleHandler(scheduleService)
	// Настраиваем маршруты через отдельную функцию в delivery слое
	router := http.SetupRouter(tokenRepo,
		authHandler,
		jwtManager,
		groupHandle,
		institutionHandler,
		subjectHandle,
		scheduleHandler,
		logger,
	)

	// Запускаем сервер
	port := ":" + cfg.ServerPort
	logger.Infof("🚀 Сервер запущен на порту %s", port)
	log.Fatal(router.Run(port))
}
