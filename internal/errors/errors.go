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
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrUnexpectedError      = errors.New("unexpected error occurred")
)
