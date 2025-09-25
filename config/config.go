package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Logger   Logger
	Bot      BotConfig
}

type BotConfig struct {
	Token     string
	WebAppURL string
}

type ServerConfig struct {
	Port       string
	Debug      bool
	Mode       string
	AppVersion string
}

type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
	PgDriver           string
}

type Logger struct {
	Development       bool
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
}

func LoadConfig(filename string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName(filename)
	v.SetConfigType("yaml")

	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./config/ymls")
	v.AddConfigPath("../config/ymls")
	v.AddConfigPath("./cmd/api")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.BindEnv("server.port", "PORT")
	v.BindEnv("server.mode", "MODE", "GIN_MODE")
	v.BindEnv("server.debug", "DEBUG")
	v.BindEnv("server.appversion", "APP_VERSION")
	v.BindEnv("postgres.postgresqlhost", "DB_HOST", "POSTGRES_HOST")
	v.BindEnv("postgres.postgresqlport", "DB_PORT", "POSTGRES_PORT")
	v.BindEnv("postgres.postgresqluser", "DB_USER", "POSTGRES_USER")
	v.BindEnv("postgres.postgresqlpassword", "DB_PASSWORD", "POSTGRES_PASSWORD")
	v.BindEnv("postgres.postgresqldbname", "DB_NAME", "POSTGRES_DB")
	v.BindEnv("postgres.postgresqlsslmode", "DB_SSLMODE", "POSTGRES_SSLMODE")
	v.BindEnv("logger.level", "LOG_LEVEL")
	v.BindEnv("logger.development", "LOG_DEVELOPMENT")
	v.BindEnv("bot.token", "BOT_TOKEN")
    v.BindEnv("bot.webappurl", "BOT_WEBAPP_URL")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using environment variables only")
			return v, nil
		}
		return nil, err
	}

	log.Printf("Using config file: %s", v.ConfigFileUsed())
	return v, nil
}

func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config
	setDefaults(v)
	err := v.Unmarshal(&c)
	if err != nil {
		log.Printf("unable to decode into struct, %v", err)
		return nil, err
	}

	return &c, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", ":8080")
	v.SetDefault("server.mode", "development")
	v.SetDefault("server.debug", false)
	v.SetDefault("server.appversion", "1.0.0")

	// Postgres defaults
	v.SetDefault("postgres.postgresqlhost", "localhost")
	v.SetDefault("postgres.postgresqlport", "15432")
	v.SetDefault("postgres.postgresqluser", "postgres")
	v.SetDefault("postgres.postgresqlpassword", "postgres")
	v.SetDefault("postgres.postgresqldbname", "postgres")
	v.SetDefault("postgres.postgresqlsslmode", false)
	v.SetDefault("postgres.pgdriver", "pgx")

	// Logger defaults
	v.SetDefault("logger.development", true)
	v.SetDefault("logger.disablecaller", false)
	v.SetDefault("logger.disablestacktrace", false)
	v.SetDefault("logger.encoding", "json")
	v.SetDefault("logger.level", "info")
}
