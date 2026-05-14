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

	// Product
	ErrProductNotFound   = errors.New("product not found")
	ErrInvalidPrice      = errors.New("price cannot be negative")
	ErrInvalidStock      = errors.New("stock cannot be negative")
	ErrInsufficientStock = errors.New("insufficient stock")

	// Order
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderEmpty            = errors.New("order must have at least one item")
	ErrInvalidQuantity       = errors.New("quantity must be greater than zero")
	ErrInvalidOrderStatus    = errors.New("invalid order status")
	ErrOrderStatusTransition = errors.New("order cannot transition to this status")
)
