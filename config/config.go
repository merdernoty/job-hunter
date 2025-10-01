package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Logger   Logger         `mapstructure:"logger"`
	Bot      BotConfig      `mapstructure:"bot"`
	MiniO    MiniOConfig    `mapstructure:"minio"`
	Jwt      JWTConfig      `mapstructure:"jwt"`
}

type JWTConfig struct {
	Secret string        `mapstructure:"jwt_secret"`
	TTL    time.Duration `mapstructure:"jwt_ttl"`
}

type MiniOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"accesskeyid"`
	SecretAccessKey string `mapstructure:"secretaccesskey"`
	BucketName      string `mapstructure:"bucketname"`
	UseSSL          bool   `mapstructure:"usessl"`
	Region          string `mapstructure:"region"`
}

type BotConfig struct {
	Token     string `mapstructure:"token"`
	WebAppURL string `mapstructure:"webappurl"`
}

type ServerConfig struct {
	Port       string `mapstructure:"port"`
	Debug      bool   `mapstructure:"debug"`
	Mode       string `mapstructure:"mode"`
	AppVersion string `mapstructure:"appversion"`
}

type PostgresConfig struct {
	PostgresqlHost     string `mapstructure:"postgresqlhost"`
	PostgresqlPort     string `mapstructure:"postgresqlport"`
	PostgresqlUser     string `mapstructure:"postgresqluser"`
	PostgresqlPassword string `mapstructure:"postgresqlpassword"`
	PostgresqlDbname   string `mapstructure:"postgresqldbname"`
	PostgresqlSSLMode  bool   `mapstructure:"postgresqlsslmode"`
	PgDriver           string `mapstructure:"pgdriver"`
}

type Logger struct {
	Development       bool   `mapstructure:"development"`
	DisableCaller     bool   `mapstructure:"disablecaller"`
	DisableStacktrace bool   `mapstructure:"disablestacktrace"`
	Encoding          string `mapstructure:"encoding"`
	Level             string `mapstructure:"level"`
}

func LoadConfig() (*viper.Viper, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults(v)

	log.Println("Using environment variables only (no config files)")
	return v, nil
}

func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config
	err := v.Unmarshal(&c)
	if err != nil {
		log.Printf("Unable to decode into struct: %v", err)
		return nil, err
	}

	log.Printf("Config loaded successfully - Server: %s, Mode: %s, Bot: %t",
		c.Server.Port, c.Server.Mode, c.Bot.Token != "")

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

	// Bot defaults
	v.SetDefault("bot.token", "")
	v.SetDefault("bot.webappurl", "")

	// MiniO defaults
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.accesskeyid", "minioadmin")
	v.SetDefault("minio.secretaccesskey", "minioadmin")
	v.SetDefault("minio.bucketname", "job-hunter-files")
	v.SetDefault("minio.usessl", false)
	v.SetDefault("minio.region", "us-east-1")

	// JWT defaults
	v.SetDefault("jwt.jwt_secret", "sadasdasd123sd")
	v.SetDefault("jwt.jwt_ttl", 24*time.Hour)
}
