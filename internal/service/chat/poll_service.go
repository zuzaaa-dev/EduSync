package chat

import (
	"EduSync/internal/delivery/http/chat/dto"
	"EduSync/internal/delivery/ws"
	"EduSync/internal/service"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"

	"EduSync/internal/domain/chat"
	"EduSync/internal/repository"
)

type pollService struct {
	repo     repository.PollRepository
	chatRepo repository.ChatRepository
	log      *logrus.Logger
	hub      *ws.Hub
}

func NewPollService(
	repo repository.PollRepository,
	chatRepo repository.ChatRepository,
	logger *logrus.Logger,
	hub *ws.Hub,
) service.PollService {
	return &pollService{
		repo:     repo,
		chatRepo: chatRepo,
		log:      logger,
		hub:      hub,
	}
}

func (s *pollService) CreatePoll(ctx context.Context, userID, chatID int, question string, options []string) (int, error) {
	// 1) Проверить, что userID — владелец чата
	ok, err := s.chatRepo.IsOwner(ctx, chatID, userID)
	if err != nil {
		s.log.Errorf("CreatePoll: IsOwner error: %v", err)
		return 0, fmt.Errorf("internal error")
	}
	if !ok {
		return 0, fmt.Errorf("permission denied")
	}
	// 2) Создать запись в polls
	poll := &chat.Poll{
		ChatID:    chatID,
		Question:  question,
		CreatedAt: time.Now(),
	}
	pollID, err := s.repo.CreatePoll(ctx, poll)
	if err != nil {
		s.log.Errorf("CreatePoll: %v", err)
		return 0, fmt.Errorf("cannot create poll")
	}
	// 3) Добавить варианты
	for _, text := range options {
		opt := &chat.Option{PollID: pollID, Text: text}
		if _, err := s.repo.CreateOption(ctx, opt); err != nil {
			s.log.Errorf("CreatePoll: CreateOption: %v", err)
			_ = s.repo.DeletePoll(ctx, pollID)
		}
	}

	summary, _ := s.buildPollSummary(ctx, pollID)

	room := fmt.Sprintf("chat_%d", chatID)
	s.hub.Broadcast(room, "poll:new", summary)

	return pollID, nil
}

func (s *pollService) DeletePoll(ctx context.Context, userID, pollID int) error {
	// найдём опрос, чтобы получить chatID
	poll, err := s.repo.GetPollByID(ctx, pollID)
	if err != nil {
		s.log.Errorf("DeletePoll: GetPollByID: %v", err)
		return fmt.Errorf("internal error")
	}
	if poll == nil {
		return fmt.Errorf("not found")
	}
	ok, err := s.chatRepo.IsOwner(ctx, poll.ChatID, userID)
	if err != nil {
		s.log.Errorf("DeletePoll: IsOwner: %v", err)
		return fmt.Errorf("internal error")
	}
	if !ok {
		return fmt.Errorf("permission denied")
	}
	if err := s.repo.DeletePoll(ctx, pollID); err != nil {
		s.log.Errorf("DeletePoll: %v", err)
		return fmt.Errorf("cannot delete poll")
	}
	room := fmt.Sprintf("chat_%d", poll.ChatID)
	s.hub.Broadcast(room, "poll:delete", map[string]int{"id": pollID})

	return nil
}

func (s *pollService) Vote(ctx context.Context, userID, pollID, optionID int) error {
	// 1) Получить вариант
	opt, err := s.repo.GetOptionByID(ctx, optionID)
	if err != nil {
		s.log.Errorf("Vote: GetOptionByID error: %v", err)
		return fmt.Errorf("internal error")
	}
	if opt == nil {
		return fmt.Errorf("not found")
	}
	// 2) Проверить, что option действительно принадлежит poll из URL
	if opt.PollID != pollID {
		return fmt.Errorf("invalid option")
	}
	// 3) Проверить, что пользователь участник чата этого опроса
	poll, err := s.repo.GetPollByID(ctx, pollID)
	if err != nil {
		s.log.Errorf("Vote: GetPollByID error: %v", err)
		return fmt.Errorf("internal error")
	}
	ok, err := s.chatRepo.IsParticipant(ctx, poll.ChatID, userID)
	if err != nil {
		s.log.Errorf("Vote: IsParticipant error: %v", err)
		return fmt.Errorf("internal error")
	}
	if !ok {
		return fmt.Errorf("permission denied")
	}
	// 4) Сохранить голос
	if err := s.repo.AddVote(ctx, &chat.Vote{UserID: userID, PollOptionID: optionID}); err != nil {
		s.log.Errorf("Vote: AddVote error: %v", err)
		return fmt.Errorf("cannot vote")
	}
	return nil
}

func (s *pollService) ListPolls(ctx context.Context, userID, chatID, limit, offset int) ([]*dto.PollSummary, error) {
	// проверяем, что user участник чата
	ok, err := s.chatRepo.IsParticipant(ctx, chatID, userID)
	if err != nil {
		s.log.Errorf("ListPolls: IsParticipant: %v", err)
		return nil, fmt.Errorf("internal error")
	}
	if !ok {
		return nil, fmt.Errorf("permission denied")
	}

	polls, err := s.repo.ListPollsByChat(ctx, chatID, limit, offset)
	if err != nil {
		s.log.Errorf("ListPolls: %v", err)
		return nil, fmt.Errorf("cannot list polls")
	}

	var out []*dto.PollSummary
	for _, p := range polls {
		opts, err := s.repo.ListOptions(ctx, p.ID)
		if err != nil {
			s.log.Errorf("ListPolls: ListOptions: %v", err)
			return nil, fmt.Errorf("internal error")
		}
		var ows []dto.OptionWithCount
		for _, o := range opts {
			cnt, _ := s.repo.CountVotes(ctx, o.ID)
			ows = append(ows, dto.OptionWithCount{
				ID:    o.ID,
				Text:  o.Text,
				Votes: cnt,
			})
		}
		out = append(out, &dto.PollSummary{
			ID:        p.ID,
			Question:  p.Question,
			CreatedAt: p.CreatedAt,
			Options:   ows,
		})
	}
	return out, nil
}

func (s *pollService) Unvote(ctx context.Context, userID, pollID, optionID int) error {
	if err := s.repo.RemoveVote(ctx, userID, optionID); err != nil {
		s.log.Errorf("Unvote: %v", err)
		return fmt.Errorf("cannot unvote")
	}
	return nil
}

func (s *pollService) buildPollSummary(ctx context.Context, pollID int) (*dto.PollSummary, error) {
	poll, err := s.repo.GetPollByID(ctx, pollID)
	if err != nil {
		s.log.Errorf("buildPollSummary: GetPollByID: %v", err)
		return nil, fmt.Errorf("cannot load poll")
	}
	if poll == nil {
		return nil, fmt.Errorf("poll not found")
	}

	opts, err := s.repo.ListOptions(ctx, pollID)
	if err != nil {
		s.log.Errorf("buildPollSummary: ListOptions: %v", err)
		return nil, fmt.Errorf("cannot load options")
	}

	summary := &dto.PollSummary{
		ID:        poll.ID,
		Question:  poll.Question,
		CreatedAt: poll.CreatedAt,
		Options:   make([]dto.OptionWithCount, 0, len(opts)),
	}
	for _, opt := range opts {
		cnt, err := s.repo.CountVotes(ctx, opt.ID)
		if err != nil {
			s.log.Errorf("buildPollSummary: CountVotes(opt %d): %v", opt.ID, err)
			cnt = 0
		}
		summary.Options = append(summary.Options, dto.OptionWithCount{
			ID:    opt.ID,
			Text:  opt.Text,
			Votes: cnt,
		})
	}

	return summary, nil
}
