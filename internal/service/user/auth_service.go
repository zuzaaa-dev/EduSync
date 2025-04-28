package user

import (
	domainUser "EduSync/internal/domain/user"
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"

	"EduSync/internal/util"
	"golang.org/x/crypto/bcrypt"
)

// AuthService управляет процессами регистрации, авторизации и логаута.
type AuthService struct {
	userRepo          repository.UserRepository
	studentRepo       repository.StudentRepository
	teacherRepo       repository.TeacherRepository
	tokenRepo         repository.TokenRepository
	instEmailMaskRepo repository.EmailMaskRepository
	jwtManager        *util.JWTManager
	log               *logrus.Logger
}

// NewAuthService создает новый экземпляр AuthService.
func NewAuthService(userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	teacherRepo repository.TeacherRepository,
	tokenRepo repository.TokenRepository,
	instEmailMaskRepo repository.EmailMaskRepository,
	jwtManager *util.JWTManager,
	log *logrus.Logger) service.UserService {
	return &AuthService{userRepo: userRepo,
		studentRepo:       studentRepo,
		teacherRepo:       teacherRepo,
		tokenRepo:         tokenRepo,
		instEmailMaskRepo: instEmailMaskRepo,
		jwtManager:        jwtManager,
		log:               log,
	}
}

// Register создает нового пользователя.
func (s *AuthService) Register(ctx context.Context, user domainUser.CreateUser) (int, error) {
	s.log.Infof("Регистрация пользователя с email: %s", user.Email)
	if user.IsTeacher {
		emailMask, err := user.EmailMask()
		if err != nil {
			s.log.Errorf("Ошибка проверки маски: %v", err)
			return 0, err
		}
		maskValid, err := s.instEmailMaskRepo.ByEmailMask(ctx, emailMask)
		if err != nil {
			s.log.Errorf("Ошибка получения маски: %v", err)
			return 0, err
		}
		if maskValid == nil {
			s.log.Errorf("Ошибка, маски не сущетсвует: %v", err)
			return 0, errors.New("email не поддерживается учреждением")
		}
	}

	// Проверяем, существует ли пользователь с таким email.
	existingUser, err := s.userRepo.ByEmail(ctx, user.Email)
	if err != nil {
		s.log.Errorf("Ошибка поиска пользователя: %v", err)
		return 0, err
	}
	if existingUser != nil {
		s.log.Errorf("пользователь с email: %v уже существует", user.Email)
		return 0, domainUser.ErrUserAlreadyExists
	}

	// Хешируем пароль.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// Начинаем транзакцию
	tx, err := s.userRepo.(interface {
		BeginTx(ctx context.Context) (*sql.Tx, error)
	}).BeginTx(ctx)
	if err != nil {
		return 0, fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	// Откат транзакции при ошибке
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Создаем пользователя.
	userID, err := s.userRepo.(interface {
		Create(ctx context.Context, tx *sql.Tx, user *domainUser.User) (int, error)
	}).Create(ctx, tx, user.ConvertToUser(hashedPassword))
	// Если студент, сохраняем данные в таблице студентов
	if !user.IsTeacher {
		err = s.studentRepo.(interface {
			Create(ctx context.Context, tx *sql.Tx, userID, institutionID, groupID int) error
		}).Create(ctx, tx, userID, user.InstitutionID, user.GroupID)
	} else {
		err = s.teacherRepo.(interface {
			Create(ctx context.Context, tx *sql.Tx, userID, institutionID int) error
		}).Create(ctx, tx, userID, user.InstitutionID)
	}
	if err != nil {
		s.log.Errorf("Ошибка создания пользователя: %v", err)
		return 0, err
	}
	// Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		s.log.Errorf("ошибка коммита транзакции: %v", err)
		return 0, domainUser.ErrInvalidCredentials
	}
	return userID, err
}

// Login выполняет авторизацию пользователя: сравнивает пароль и генерирует токены.
func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (string, string, error) {
	s.log.Infof("Регистрация пользователя с email: %s", email)
	user, err := s.userRepo.ByEmail(ctx, email)
	if err != nil {
		s.log.Errorf("Ошибка поиска пользователя: %v", err)
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("неверный email или пароль")
	}

	// Сравниваем пароль.
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		s.log.Errorf("Ошибка хэширования %v", err)
		return "", "", errors.New("неверный email или пароль")
	}

	// Получаем информацию об учебном заведении и группе
	var institutionId, groupId int

	if user.IsTeacher {
		teacher, err := s.teacherRepo.ByUserID(ctx, user.ID)
		if err != nil {
			s.log.Errorf("Ошибка получения информации о преподавателе: %v", err)
			return "", "", err
		}
		institutionId = teacher.InstitutionID
		groupId = 0 // У преподавателя нет группы
	} else {
		student, err := s.studentRepo.ByUserID(ctx, user.ID)
		if err != nil {
			s.log.Errorf("Ошибка получения информации о студенте: %v", err)
			return "", "", err
		}
		institutionId = student.InstitutionID
		groupId = student.GroupID
	}

	// Удаляем предыдущие токены для пользователя, чтобы не было дублирования.
	if err := s.tokenRepo.DeleteForUser(ctx, user.ID); err != nil {
		s.log.Errorf("Ошибка удаления токенов: %v", err)
		return "", "", err
	}

	// Генерируем access token.
	accessToken, err := s.jwtManager.GenerateJWT(user.ID, user.IsTeacher, user.Email, user.FullName, institutionId, groupId, time.Hour)
	if err != nil {
		s.log.Errorf("Ошибка генерации токена: %v", err)
		return "", "", err
	}

	// Генерируем refresh token.
	refreshToken, err := s.jwtManager.GenerateJWT(user.ID, user.IsTeacher, user.Email, user.FullName, institutionId, groupId, 24*time.Hour)
	if err != nil {
		s.log.Errorf("Ошибка генерации рефреш токена: %v", err)
		return "", "", err
	}

	// Сохраняем токены в БД.
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := s.tokenRepo.Save(ctx, user.ID, accessToken, refreshToken, userAgent, ipAddress, expiresAt); err != nil {
		s.log.Errorf("Ошибка сохранения токенов: %v", err)
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, u domainUser.UpdateUser, userAgent string, ipAddress string) (string, string, error) {
	// 1) начинаем tx
	tx, err := s.userRepo.BeginTx(ctx)
	if err != nil {
		s.log.Errorf("BeginTx: %v", err)
		return "", "", fmt.Errorf("не удалось начать транзакцию")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 2) проверяем, что пользователь существует
	existing, err := s.userRepo.ByID(ctx, u.ID)
	if err != nil {
		return "", "", err
	}
	if existing == nil {
		return "", "", fmt.Errorf("пользователь не найден")
	}

	// 3) обновляем таблицу users
	updateUser := &domainUser.User{
		ID:       u.ID,
		FullName: existing.FullName,
	}
	if u.FullName != nil {
		updateUser.FullName = *u.FullName
	}
	if err = s.userRepo.Update(ctx, tx, updateUser); err != nil {
		s.log.Errorf("userRepo.Update: %v", err)
		return "", "", fmt.Errorf("не удалось обновить пользователя")
	}

	// 4) обновляем students/teachers
	if existing.IsTeacher {
		if u.InstitutionID != nil {
			if err = s.teacherRepo.Update(ctx, tx, u.ID, *u.InstitutionID); err != nil {
				s.log.Errorf("teacherRepo.Update: %v", err)
				return "", "", fmt.Errorf("не удалось обновить данные преподавателя")
			}
		}
	} else {
		// студент
		stu, err2 := s.studentRepo.ByUserID(ctx, u.ID)
		if err2 != nil {
			return "", "", err2
		}
		instID := stu.InstitutionID
		grpID := stu.GroupID
		if u.InstitutionID != nil {
			instID = *u.InstitutionID
		}
		if u.GroupID != nil {
			grpID = *u.GroupID
		}
		if err = s.studentRepo.Update(ctx, tx, u.ID, instID, grpID); err != nil {
			s.log.Errorf("studentRepo.Update: %v", err)
			return "", "", fmt.Errorf("не удалось обновить данные студента")
		}
	}

	// 5) фиксим транзакцию
	if err = tx.Commit(); err != nil {
		s.log.Errorf("tx.Commit: %v", err)
		return "", "", fmt.Errorf("не удалось сохранить изменения")
	}

	// 6) достаём всё для новой пачки claim’ов
	user, err2 := s.userRepo.ByID(ctx, u.ID)
	if err2 != nil {
		return "", "", err2
	}
	var institutionID, groupID int
	if user.IsTeacher {
		t, _ := s.teacherRepo.ByUserID(ctx, u.ID)
		institutionID = t.InstitutionID
	} else {
		st, _ := s.studentRepo.ByUserID(ctx, u.ID)
		institutionID = st.InstitutionID
		groupID = st.GroupID
	}

	// 7) генерим новые токены
	newAccess, err := s.jwtManager.GenerateJWT(
		user.ID,
		user.IsTeacher,
		user.Email,
		user.FullName,
		institutionID,
		groupID,
		time.Hour,
	)
	if err != nil {
		s.log.Errorf("GenerateTokens: %v", err)
		return "", "", fmt.Errorf("не удалось обновить токен")
	}
	newRefresh, err := s.jwtManager.GenerateJWT(
		user.ID,
		user.IsTeacher,
		user.Email,
		user.FullName,
		institutionID,
		groupID,
		24*time.Hour,
	)

	// 8) чистим старые и сохраняем новые в БД
	if err = s.tokenRepo.DeleteForUser(ctx, user.ID); err != nil {
		s.log.Errorf("tokenRepo.DeleteForUser: %v", err)
		return "", "", err
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err = s.tokenRepo.Save(ctx, user.ID, newAccess, newRefresh, userAgent, ipAddress, expiresAt); err != nil {
		s.log.Errorf("tokenRepo.Save: %v", err)
		return "", "", err
	}

	return newAccess, newRefresh, nil
}

// Logout отзываёт токен: удаляет токен из БД.
func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	return s.tokenRepo.Revoke(ctx, accessToken)
}

// RefreshToken обновляет access-токен, если refresh-токен валиден.
func (s *AuthService) RefreshToken(ctx context.Context, inputRefreshToken, userAgent, ipAddress string) (string, string, error) {
	// Проверяем, существует ли refresh-токен в БД (используем отдельную функцию для refresh-токенов)
	exists, err := s.tokenRepo.IsRefreshValid(ctx, inputRefreshToken)
	if err != nil || !exists {
		s.log.Errorf("Ошибка провки токенов или токена нет в БД: %v", err)
		return "", "", errors.New("недействительный refresh-токен")
	}

	claims, err := s.jwtManager.ParseJWT(inputRefreshToken, s.log)
	if err != nil {
		return "", "", errors.New("недействительный или просроченный refresh-токен")
	}

	// Генерируем новый access-токен
	accessToken, err := s.jwtManager.GenerateJWT(claims.ID, claims.IsTeacher, claims.Email, claims.FullName, claims.InstitutionId, claims.GroupId, time.Hour)
	if err != nil {
		s.log.Errorf("Ошибка генерации токенов: %v", err)
		return "", "", err
	}

	// Генерируем новый refresh-токен
	newRefreshToken, err := s.jwtManager.GenerateJWT(claims.ID, claims.IsTeacher, claims.Email, claims.FullName, claims.InstitutionId, claims.GroupId, 7*24*time.Hour)
	if err != nil {
		s.log.Errorf("Ошибка генерации рефреш токена: %v", err)
		return "", "", err
	}

	// Обновляем токены в БД: удаляем старые токены и сохраняем новые с актуальными значениями userAgent и ipAddress
	if err := s.tokenRepo.DeleteForUser(ctx, claims.ID); err != nil {
		s.log.Errorf("Ошибка удаления токенов: %v", err)
		return "", "", err
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := s.tokenRepo.Save(ctx, claims.ID, accessToken, newRefreshToken, userAgent, ipAddress, expiresAt); err != nil {
		s.log.Errorf("Ошибка сохранения токена: %v", err)
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// FindTeacherByName ищет преподавателя по строке с инициалами вида "Коноплев А.А."
func (s *AuthService) FindTeacherByName(ctx context.Context, teacherStr string) (*domainUser.User, error) {
	// Разбиваем входящую строку по пробелу
	parts := strings.Fields(teacherStr)
	if len(parts) < 2 {
		return nil, fmt.Errorf("недостаточно данных в строке: %s", teacherStr)
	}
	surname := parts[0]
	initialsProvided := strings.ReplaceAll(parts[1], ".", "") // Убираем точки, получаем, например, "АА"

	// Получаем преподавателей по фамилии
	teachers, err := s.teacherRepo.BySurname(ctx, surname)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения преподавателей: %w", err)
	}

	// Перебираем полученных преподавателей
	for _, t := range teachers {
		nameParts := strings.Fields(t.FullName)
		if len(nameParts) < 2 {
			continue
		}
		// Фамилия должна совпадать
		if !strings.EqualFold(nameParts[0], surname) {
			continue
		}

		var initials string
		for _, part := range nameParts[1:] {
			if len(part) > 0 {
				initials += part[0:2]
			}
		}

		if initials == initialsProvided {
			return t, nil
		}
	}

	return nil, fmt.Errorf("преподаватель с инициалами %s не найден", teacherStr)
}
