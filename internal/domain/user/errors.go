package user

import "errors"

var (
	// ErrUserAlreadyExists возвращается, если пользователь с заданным email уже существует.
	ErrUserAlreadyExists = errors.New("пользователь с таким email уже существует")
	// ErrRegistrationFailed – общее сообщение об ошибке регистрации.
	ErrRegistrationFailed = errors.New("ошибка регистрации пользователя")
	// ErrInvalidCredentials – общее сообщение об ошибке аутентификации.
	ErrInvalidCredentials = errors.New("неверные учетные данные")
)
