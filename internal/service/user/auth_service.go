package user

import (
	domainUser "EduSync/internal/domain/user"
	"EduSync/internal/repository"
	"errors"
	"time"

	"EduSync/internal/util"
	"golang.org/x/crypto/bcrypt"
)

// AuthService управляет процессами регистрации, авторизации и логаута.
type AuthService struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.TokenRepository
	jwtManager *util.JWTManager
}

// NewAuthService создает новый экземпляр AuthService.
func NewAuthService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, jwtManager *util.JWTManager) *AuthService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, jwtManager: jwtManager}
}

// Register создает нового пользователя.
func (s *AuthService) Register(user domainUser.CreateUser) (int, error) {
	// Проверяем, существует ли пользователь с таким email.
	existingUser, err := s.userRepo.GetUserByEmail(user.Email)
	if err != nil {
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
	userID, err := s.userRepo.CreateUser(user.ConvertToUser(&hashedPassword))
	return userID, err
}

// Login выполняет авторизацию пользователя: сравнивает пароль и генерирует токены.
func (s *AuthService) Login(email, password, userAgent, ipAddress string) (string, string, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("неверный email или пароль")
	}

	// Сравниваем пароль.
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", errors.New("неверный email или пароль")
	}

	// Удаляем предыдущие токены для пользователя, чтобы не было дублирования.
	if err := s.tokenRepo.DeleteTokensForUser(user.ID); err != nil {
		return "", "", err
	}

	// Генерируем access token.
	accessToken, err := s.jwtManager.GenerateJWT(user.ID, user.IsTeacher, user.Email, user.FullName, time.Hour)
	if err != nil {
		return "", "", err
	}

	// Генерируем refresh token.
	refreshToken, err := s.jwtManager.GenerateJWT(user.ID, user.IsTeacher, user.Email, user.FullName, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	// Сохраняем токены в БД.
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := s.tokenRepo.SaveToken(user.ID, accessToken, refreshToken, userAgent, ipAddress, expiresAt); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Logout отзываёт токен: удаляет токен из БД.
func (s *AuthService) Logout(accessToken string) error {
	return s.tokenRepo.RevokeToken(accessToken)
}

// RefreshToken обновляет access-токен, если refresh-токен валиден.
func (s *AuthService) RefreshToken(inputRefreshToken, userAgent, ipAddress string) (string, string, error) {
	// Проверяем, существует ли refresh-токен в БД (используем отдельную функцию для refresh-токенов)
	exists, err := s.tokenRepo.IsRefreshTokenValid(inputRefreshToken)
	if err != nil || !exists {
		return "", "", errors.New("недействительный refresh-токен")
	}

	claims, err := s.jwtManager.ParseJWT(inputRefreshToken)
	if err != nil {
		return "", "", errors.New("недействительный или просроченный refresh-токен")
	}

	// Генерируем новый access-токен
	accessToken, err := s.jwtManager.GenerateJWT(claims.ID, claims.IsTeacher, claims.Email, claims.FullName, time.Hour)
	if err != nil {
		return "", "", err
	}

	// Генерируем новый refresh-токен
	newRefreshToken, err := s.jwtManager.GenerateJWT(claims.ID, claims.IsTeacher, claims.Email, claims.FullName, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	// Обновляем токены в БД: удаляем старые токены и сохраняем новые с актуальными значениями userAgent и ipAddress
	if err := s.tokenRepo.DeleteTokensForUser(claims.ID); err != nil {
		return "", "", err
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := s.tokenRepo.SaveToken(claims.ID, accessToken, newRefreshToken, userAgent, ipAddress, expiresAt); err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}
