package chat

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrAlreadyFavorited = errors.New("already favorited")
	ErrNotFavorited     = errors.New("not favorited")
)
