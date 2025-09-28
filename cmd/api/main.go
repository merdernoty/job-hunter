package main

import (
	"github.com/merdernoty/job-hunter/app"
	"github.com/merdernoty/job-hunter/config"
	"github.com/merdernoty/job-hunter/internal/bot"
	user "github.com/merdernoty/job-hunter/internal/users"
	"github.com/merdernoty/job-hunter/pkg/db/postgres"
	"github.com/merdernoty/job-hunter/pkg/env"
	httpPkg "github.com/merdernoty/job-hunter/pkg/http"
	"github.com/merdernoty/job-hunter/pkg/logger"
	"github.com/merdernoty/job-hunter/pkg/storage"
	"github.com/merdernoty/job-hunter/pkg/telegram"
	"go.uber.org/fx"
)

func init() {
	env.Load()
}

func main() {
	fx.New(
		config.Module,
		postgres.Module,
		app.Module,
		httpPkg.Module,
		logger.Module,
		bot.Module,
		storage.Module,
		telegram.Module,
		user.Module,
	).Run()
}
