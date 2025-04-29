package favorite

import "errors"

var (
	ErrAlreadyFavorited = errors.New("file already in favorites")
	ErrFavoriteNotFound = errors.New("favorite not found")
)
