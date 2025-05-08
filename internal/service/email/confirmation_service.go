package email

import (
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"time"
)

type confirmationService struct {
	repo      repository.EmailConfirmationsRepository
	userRepo  repository.UserRepository
	emailSvc  service.EmailService
	throttle  time.Duration
	codeTTL   time.Duration
	templates map[string]string
	baseURL   string
}

func NewConfirmationService(
	repo repository.EmailConfirmationsRepository,
	userRepo repository.UserRepository,
	emailSvc service.EmailService,
	baseURL string,
) service.ConfirmationService {
	return &confirmationService{
		repo:     repo,
		userRepo: userRepo,
		emailSvc: emailSvc,
		throttle: 2 * time.Minute,
		codeTTL:  2 * time.Hour,
		baseURL:  baseURL,
		templates: map[string]string{
			"register":       "Подтверждение регистрации",
			"reset_password": "Сброс пароля",
			"delete_account": "Подтверждение удаления аккаунта",
		},
	}
}

func (s *confirmationService) genCode() string {
	n := rand.Intn(1e6)
	return fmt.Sprintf("%06d", n)
}

func (s *confirmationService) RequestCode(ctx context.Context, email, action string) error {
	user, err := s.userRepo.ByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	ok, err := s.repo.CanSendNew(ctx, user.ID, action, s.throttle)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("код уже отправлялся недавно")
	}
	code := s.genCode()
	expires := time.Now().Add(s.codeTTL)
	if err := s.repo.Create(ctx, user.ID, action, code, expires); err != nil {
		return err
	}
	subj := s.templates[action]

	var body string
	if action == "register" {
		// Формируем ссылку на активацию
		activationLink := fmt.Sprintf(
			"%s/api/confirm/activate?user_id=%d&code=%s",
			s.baseURL, user.ID, code,
		)
		body = fmt.Sprintf(
			"Чтобы активировать аккаунт, перейдите по ссылке:\n\n%s\n\nСсылка действительна %d часов.",
			activationLink, int(s.codeTTL.Hours()),
		)
	} else {
		// Для остальных действий остаётся код
		body = fmt.Sprintf("Ваш код подтверждения: %s\nОн действителен %d часов.", code, int(s.codeTTL.Hours()))
	}

	return s.emailSvc.SendCode(ctx, email, subj, body)
}

func (s *confirmationService) VerifyCode(ctx context.Context, action, code string, userID *int) error {
	if userID == nil {
		newUserID, err := s.repo.GetByActionCode(ctx, action, code)
		if err != nil {
			return err
		}
		userID = &newUserID
	}

	valid, err := s.repo.GetValid(ctx, *userID, action, code)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("код неверен или просрочен")
	}

	if action != "reset_password" {
		if err := s.repo.MarkUsed(ctx, *userID, action, code); err != nil {
			return err
		}
	}
	switch action {
	case "register":
		if err := s.userRepo.Activate(ctx, *userID); err != nil {
			return fmt.Errorf("не удалось активировать пользователя")
		}
	}
	return nil
}

func (s *confirmationService) UserIDByCode(ctx context.Context, action, code string) (int, error) {
	// допустим, в email_confirmations есть user_id
	userID, err := s.repo.GetByActionCode(ctx, action, code)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (s *confirmationService) ResetPassword(ctx context.Context, code, newPassword string) error {
	// 1) найти userID по коду
	userID, err := s.repo.GetByActionCode(ctx, "reset_password", code)
	if err != nil {
		return fmt.Errorf("invalid or expired code")
	}
	// 2) пометить код использованным
	if err := s.repo.MarkUsed(ctx, userID, "reset_password", code); err != nil {
		return err
	}
	// 3) обновить пароль
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, userID, string(hashed))
}
