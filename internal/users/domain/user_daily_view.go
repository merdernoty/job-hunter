package domain

import (
	"github.com/google/uuid"
	"time"
)

type UserDailyView struct {
	ViewerID    uuid.UUID `json:"viewer_id" db:"viewer_id"`
	ShownUserID uuid.UUID `json:"shown_user_id" db:"shown_user_id"`
	ViewDate    time.Time `json:"view_date" db:"view_date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type UserDailyViewRepository interface {
	Create(view *UserDailyView) error
	GetTodaysDailyUser(viewerID uuid.UUID) (*User, error)
	GetTodaysShownUsers(viewerID uuid.UUID) ([]uuid.UUID, error)
	IsUserShownToday(viewerID, shownUserID uuid.UUID) (bool, error)
	CleanupOldRecords(daysToKeep int) error
}
