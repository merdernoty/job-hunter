package postgres

import (
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/merdernoty/job-hunter/config"
	"github.com/merdernoty/job-hunter/pkg/logger"
)

const (
	maxOpenConns    = 10
	connMaxLifetime = 0               
	maxIdleConns    = 5
	connMaxIdleTime = 30 * time.Minute
)

func NewPsqlDB(c *config.Config, log logger.Logger) (*sqlx.DB, error) {
	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=30",
		c.Postgres.PostgresqlHost,
		c.Postgres.PostgresqlPort,
		c.Postgres.PostgresqlUser,
		c.Postgres.PostgresqlPassword, 
		c.Postgres.PostgresqlDbname,
	)
	
	log.Infof("Attempting to connect to PostgreSQL: %s:%s/%s", 
		c.Postgres.PostgresqlHost, 
		c.Postgres.PostgresqlPort, 
		c.Postgres.PostgresqlDbname)
	
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		log.Errorf("Failed to create database connection: %v", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	log.Info("Database connection created successfully")

	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(connMaxIdleTime)
	
	log.Infof("Database pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v, MaxIdleTime=%v", 
		maxOpenConns, maxIdleConns, connMaxLifetime, connMaxIdleTime)
	
	if err := db.Ping(); err != nil {
		log.Errorf("Failed to ping database: %v", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	log.Info("Database ping successful - connection established")
	
	stats := db.Stats()
	log.Infof("Initial database pool stats: OpenConnections=%d, InUse=%d, Idle=%d",
		stats.OpenConnections, stats.InUse, stats.Idle)

	return db, nil
}