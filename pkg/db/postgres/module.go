package postgres

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/merdernoty/job-hunter/config"
	"github.com/merdernoty/job-hunter/pkg/logger"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewPsqlDB),
	fx.Invoke(RunMigrations),
)

func RunMigrations(lc fx.Lifecycle, db *sqlx.DB, cfg *config.Config, log logger.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Running migrations...")
			if err := AutoMigrate(db, cfg, log); err != nil {
				return err
			}
			log.Info("Migrations completed successfully")
			return nil
		},
	})
}
