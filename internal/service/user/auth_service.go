package user

import (
	domainUser "EduSync/internal/domain/user"
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"time"

	"EduSync/internal/util"
	"golang.org/x/crypto/bcrypt"
)

// AuthService управляет процессами регистрации, авторизации и логаута.
type AuthService struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.TokenRepository
	jwtManager *util.JWTManager
	log        *logrus.Logger
}

// NewAuthService создает новый экземпляр AuthService.
func NewAuthService(userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwtManager *util.JWTManager,
	log *logrus.Logger) service.UserService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, jwtManager: jwtManager, log: log}
}

// Register создает нового пользователя.
func (s *AuthService) Register(ctx context.Context, user domainUser.CreateUser) (int, error) {
	s.log.Infof("Регистрация пользователя с email: %s", user.Email)
	// Проверяем, существует ли пользователь с таким email.
	existingUser, err := s.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		s.log.Errorf("Ошибка поиска пользователя: %v", err)
		return 0, err
	}
	if existingUser != nil {
		return 0, errors.New("пользователь с таким email уже существует")
	}

	// Хешируем пароль.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// Создаем пользователя.
	userID, err := s.userRepo.CreateUser(ctx, user.ConvertToUser(&hashedPassword))
	return userID, err
}

// Login выполняет авторизацию пользователя: сравнивает пароль и генерирует токены.
func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (string, string, error) {
	s.log.Infof("Регистрация пользователя с email: %s", email)
	user, err := s.userRepo.GetUserByEmail(ctx, email)
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

	// Удаляем предыдущие токены для пользователя, чтобы не было дублирования.
	if err := s.tokenRepo.DeleteTokensForUser(ctx, user.ID); err != nil {
		s.log.Errorf("Ошибка удаления токенов: %v", err)
		return "", "", err
	}

	// Генерируем access token.
	accessToken, err := s.jwtManager.GenerateJWT(user.ID, user.IsTeacher, user.Email, user.FullName, time.Hour)
	if err != nil {
		s.log.Errorf("Ошибка генерации токена: %v", err)
		return "", "", err
	}

	// Генерируем refresh token.
	refreshToken, err := s.jwtManager.GenerateJWT(user.ID, user.IsTeacher, user.Email, user.FullName, 7*24*time.Hour)
	if err != nil {
		s.log.Errorf("Ошибка генерации рефреш токена: %v", err)
		return "", "", err
	}

	// Сохраняем токены в БД.
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := s.tokenRepo.SaveToken(ctx, user.ID, accessToken, refreshToken, userAgent, ipAddress, expiresAt); err != nil {
		s.log.Errorf("Ошибка сохранения токенов: %v", err)
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Logout отзываёт токен: удаляет токен из БД.
func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	return s.tokenRepo.RevokeToken(ctx, accessToken)
}

// RefreshToken обновляет access-токен, если refresh-токен валиден.
func (s *AuthService) RefreshToken(ctx context.Context, inputRefreshToken, userAgent, ipAddress string) (string, string, error) {
	// Проверяем, существует ли refresh-токен в БД (используем отдельную функцию для refresh-токенов)
	exists, err := s.tokenRepo.IsRefreshTokenValid(ctx, inputRefreshToken)
	if err != nil || !exists {
		s.log.Errorf("Ошибка провки токенов или токена нет в БД: %v", err)
		return "", "", errors.New("недействительный refresh-токен")
	}

	claims, err := s.jwtManager.ParseJWT(inputRefreshToken)
	if err != nil {
		return "", "", errors.New("недействительный или просроченный refresh-токен")
	}

	// Генерируем новый access-токен
	accessToken, err := s.jwtManager.GenerateJWT(claims.ID, claims.IsTeacher, claims.Email, claims.FullName, time.Hour)
	if err != nil {
		s.log.Errorf("Ошибка генерации токенов: %v", err)
		return "", "", err
	}

	// Генерируем новый refresh-токен
	newRefreshToken, err := s.jwtManager.GenerateJWT(claims.ID, claims.IsTeacher, claims.Email, claims.FullName, 7*24*time.Hour)
	if err != nil {
		s.log.Errorf("Ошибка генерации рефреш токена: %v", err)
		return "", "", err
	}

	// Обновляем токены в БД: удаляем старые токены и сохраняем новые с актуальными значениями userAgent и ipAddress
	if err := s.tokenRepo.DeleteTokensForUser(ctx, claims.ID); err != nil {
		s.log.Errorf("Ошибка удаления токенов: %v", err)
		return "", "", err
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := s.tokenRepo.SaveToken(ctx, claims.ID, accessToken, newRefreshToken, userAgent, ipAddress, expiresAt); err != nil {
		s.log.Errorf("Ошибка сохранения токена: %v", err)
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}
