package email

import (
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"fmt"
	"math/rand"
	"time"
)

type confirmationService struct {
	repo      repository.EmailConfirmationsRepository
	emailSvc  service.EmailService
	throttle  time.Duration
	codeTTL   time.Duration
	templates map[string]string
}

func NewConfirmationService(repo repository.EmailConfirmationsRepository, emailSvc service.EmailService) service.ConfirmationService {
	return &confirmationService{
		repo:     repo,
		emailSvc: emailSvc,
		throttle: 2 * time.Minute,
		codeTTL:  2 * time.Hour,
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

func (s *confirmationService) RequestCode(ctx context.Context, userID int, email, action string) error {
	ok, err := s.repo.CanSendNew(ctx, userID, action, s.throttle)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("код уже отправлялся недавно")
	}
	code := s.genCode()
	expires := time.Now().Add(s.codeTTL)
	if err := s.repo.Create(ctx, userID, action, code, expires); err != nil {
		return err
	}
	body := fmt.Sprintf("Ваш код подтверждения: %s\nОн действителен %d часов.", code, int(s.codeTTL.Hours()))
	subj := s.templates[action]
	return s.emailSvc.SendCode(ctx, email, subj, body)
}

func (s *confirmationService) VerifyCode(ctx context.Context, userID int, action, code string) error {
	valid, err := s.repo.GetValid(ctx, userID, action, code)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("код неверен или просрочен")
	}
	// помечаем использованным
	if err := s.repo.MarkUsed(ctx, userID, action, code); err != nil {
		return err
	}
	return nil
}
