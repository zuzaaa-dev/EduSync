package chat

import (
	domainChat "EduSync/internal/domain/chat"
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"
)

type pollRepository struct {
	db *sql.DB
}

func NewPollRepository(db *sql.DB) repository.PollRepository {
	return &pollRepository{db}
}

func (r *pollRepository) CreatePoll(ctx context.Context, p *domainChat.Poll) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO polls (chat_id, question, created_at)
        VALUES ($1,$2,$3) RETURNING id
    `, p.ChatID, p.Question, p.CreatedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreatePoll: %w", err)
	}
	return id, nil
}

func (r *pollRepository) DeletePoll(ctx context.Context, pollID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM polls WHERE id = $1`, pollID)
	return err
}

func (r *pollRepository) CreateOption(ctx context.Context, opt *domainChat.Option) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO poll_options (poll_id, option_text)
        VALUES ($1,$2) RETURNING id
    `, opt.PollID, opt.Text).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateOption: %w", err)
	}
	return id, nil
}

func (r *pollRepository) ListOptions(ctx context.Context, pollID int) ([]*domainChat.Option, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, poll_id, option_text
        FROM poll_options
        WHERE poll_id = $1
    `, pollID)
	if err != nil {
		return nil, fmt.Errorf("ListOptions: %w", err)
	}
	defer rows.Close()

	var opts []*domainChat.Option
	for rows.Next() {
		o := new(domainChat.Option)
		if err := rows.Scan(&o.ID, &o.PollID, &o.Text); err != nil {
			return nil, fmt.Errorf("ListOptions scan: %w", err)
		}
		opts = append(opts, o)
	}
	return opts, nil
}

func (r *pollRepository) AddVote(ctx context.Context, v *domainChat.Vote) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO votes (user_id, poll_option_id)
        VALUES ($1,$2)
    `, v.UserID, v.PollOptionID)
	return err
}

func (r *pollRepository) RemoveVote(ctx context.Context, userID, optionID int) error {
	_, err := r.db.ExecContext(ctx, `
        DELETE FROM votes WHERE user_id = $1 AND poll_option_id = $2
    `, userID, optionID)
	return err
}

func (r *pollRepository) CountVotes(ctx context.Context, optionID int) (int, error) {
	var cnt int
	err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM votes WHERE poll_option_id = $1
    `, optionID).Scan(&cnt)
	return cnt, err
}

func (r *pollRepository) GetPollByID(ctx context.Context, pollID int) (*domainChat.Poll, error) {
	p := new(domainChat.Poll)
	err := r.db.QueryRowContext(ctx, `
        SELECT id, chat_id, question, created_at
        FROM polls
        WHERE id = $1
    `, pollID).Scan(&p.ID, &p.ChatID, &p.Question, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetPollByID: %w", err)
	}
	return p, nil
}

func (r *pollRepository) GetOptionByID(ctx context.Context, optionID int) (*domainChat.Option, error) {
	o := new(domainChat.Option)
	err := r.db.QueryRowContext(ctx, `
        SELECT id, poll_id, option_text
        FROM poll_options
        WHERE id = $1
    `, optionID).Scan(&o.ID, &o.PollID, &o.Text)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetOptionByID: %w", err)
	}
	return o, nil
}

func (r *pollRepository) ListPollsByChat(ctx context.Context, chatID, limit, offset int) ([]*domainChat.Poll, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, chat_id, question, created_at
        FROM polls
        WHERE chat_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ListPollsByChat: %w", err)
	}
	defer rows.Close()

	var polls []*domainChat.Poll
	for rows.Next() {
		p := new(domainChat.Poll)
		if err := rows.Scan(&p.ID, &p.ChatID, &p.Question, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("ListPollsByChat scan: %w", err)
		}
		polls = append(polls, p)
	}
	return polls, nil
}
