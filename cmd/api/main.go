package main

import (
	"github.com/merdernoty/job-hunter/app"
	"github.com/merdernoty/job-hunter/config"
	"github.com/merdernoty/job-hunter/pkg/logger"
	"github.com/merdernoty/job-hunter/pkg/db/postgres"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		postgres.Module,
		app.Module,
		logger.Module,
		// user.Module,     // TODO
		// fx.Invoke(database.Migrate), // TODO
	).Run()
}