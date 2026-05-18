package models

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"id,omitempty"`
	Username string    `json:"username,omitempty"`
	Password string    `json:"password,omitempty"`
	Email    string    `json:"email,omitempty"`
}

//TODO: Implement refresh token
type TokenResponse struct {
	Token    string      `json:"access_token"`
	Exp      int64       `json:"exp"`
	Iat      int64       `json:"iat"`
	UserInfo interface{} `json:"user,omitempty"`
}
