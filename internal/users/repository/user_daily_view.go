package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/merdernoty/job-hunter/internal/users/domain"
	"github.com/merdernoty/job-hunter/pkg/logger"
)

type userDailyViewRepository struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewUserDailyViewRepository(db *sqlx.DB, logger logger.Logger) domain.UserDailyViewRepository {
	return &userDailyViewRepository{db: db, logger: logger}
}

func (r *userDailyViewRepository) Create(view *domain.UserDailyView) error {
	query := `
		INSERT INTO user_daily_views (viewer_id, shown_user_id, view_date, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (viewer_id, shown_user_id, view_date) DO NOTHING`

	_, err := r.db.Exec(query, view.ViewerID, view.ShownUserID, view.ViewDate, view.CreatedAt)
	if err != nil {
		r.logger.Errorf("Failed to create user daily view: %v", err)
		return fmt.Errorf("failed to create user daily view")
	}

	r.logger.Infof("Created daily view: viewer %s, shown user %s, date %s",
		view.ViewerID, view.ShownUserID, view.ViewDate.Format("2006-01-02"))
	return nil
}

func (r *userDailyViewRepository) GetTodaysDailyUser(viewerID uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT u.id, u.telegram_id, u.username, u.telegram_handle, u.avatar_url, u.bio, u.created_at, u.updated_at
		FROM users u
		JOIN user_daily_views udv ON u.id = udv.shown_user_id
		WHERE udv.viewer_id = $1 AND udv.view_date = CURRENT_DATE
		ORDER BY udv.created_at ASC
		LIMIT 1`

	err := r.db.Get(&user, query, viewerID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no daily user found for today")
	}
	if err != nil {
		r.logger.Errorf("Failed to get today's daily user for viewer %s: %v", viewerID, err)
		return nil, fmt.Errorf("database error")
	}

	return &user, nil
}

func (r *userDailyViewRepository) GetTodaysShownUsers(viewerID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	query := `
		SELECT shown_user_id 
		FROM user_daily_views 
		WHERE viewer_id = $1 AND view_date = CURRENT_DATE`

	err := r.db.Select(&userIDs, query, viewerID)
	if err != nil {
		r.logger.Errorf("Failed to get today's shown users for viewer %s: %v", viewerID, err)
		return nil, fmt.Errorf("database error")
	}

	return userIDs, nil
}

func (r *userDailyViewRepository) IsUserShownToday(viewerID, shownUserID uuid.UUID) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM user_daily_views 
		WHERE viewer_id = $1 AND shown_user_id = $2 AND view_date = CURRENT_DATE`

	err := r.db.Get(&count, query, viewerID, shownUserID)
	if err != nil {
		r.logger.Errorf("Failed to check if user was shown today: %v", err)
		return false, fmt.Errorf("database error")
	}

	return count > 0, nil
}

func (r *userDailyViewRepository) CleanupOldRecords(daysToKeep int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysToKeep)

	query := `
		DELETE FROM user_daily_views 
		WHERE view_date < $1`

	result, err := r.db.Exec(query, cutoffDate)
	if err != nil {
		r.logger.Errorf("Failed to cleanup old user daily views: %v", err)
		return fmt.Errorf("failed to cleanup old records")
	}

	rowsDeleted, _ := result.RowsAffected()
	r.logger.Infof("Cleaned up %d old user daily view records (older than %s)",
		rowsDeleted, cutoffDate.Format("2006-01-02"))

	return nil
}
