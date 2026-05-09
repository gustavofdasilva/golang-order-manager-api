package error_codes

import "errors"

var (
	ErrInvalidEmail         = errors.New("invalid email")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrInvalidUsername      = errors.New("invalid username")
	ErrUserNotFound         = errors.New("user not found")
	ErrEmailAlreadyInUse    = errors.New("email already in use")
	ErrUsernameAlreadyInUse = errors.New("username already in use")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUnexpectedError      = errors.New("unexpected error occurred")
)
