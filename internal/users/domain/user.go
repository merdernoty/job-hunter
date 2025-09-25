package domain

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TelegramID int64     `json:"telegram_id" db:"telegram_id"`
	Username   string    `json:"username" db:"username"`
	AvatarURL  string    `json:"avatar_url" db:"avatar_url"`
	Bio        string    `json:"bio" db:"bio"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type TelegramAuthRequest struct {
	InitData string `json:"initData" validate:"required"`
}

type UpdateUserRequest struct {
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Username  *string `json:"username,omitempty" validate:"omitempty,max=50"`
	Bio       *string `json:"bio,omitempty" validate:"omitempty,max=500"`
}


type UserRepository interface {
	GetByID(id uuid.UUID) (*User, error)
	GetByTelegramID(telegramID int64) (*User, error)
	Create(user *User) error
	Update(id uuid.UUID, updates UpdateUserRequest) error
}

type UserService interface {
	AuthFromTelegram(initData string) (*User, string, error) // user, token, error
	GetUser(id uuid.UUID) (*User, error)
	UpdateUser(id uuid.UUID, req UpdateUserRequest) (*User, error)
}