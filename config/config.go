package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	RedisAddr            string
	JWTSecret            string
	ServerPort           string
	CORSAllowOrigins     []string
	CORSAllowMethods     []string
	CORSAllowHeaders     []string
	CORSAllowCredentials bool
	CORSExposeHeaders    string
	CORSMaxAge           string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables")
	}

	return &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "user"),
		DBPassword:           getEnv("DB_PASSWORD", "password"),
		DBName:               getEnv("DB_NAME", "warehouse"),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:            getEnv("JWT_SECRET", "Ult_wms"),
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		CORSAllowOrigins:     splitEnv("CORS_ALLOW_ORIGINS", "*"),
		CORSAllowMethods:     splitEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
		CORSAllowHeaders:     splitEnv("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization"),
		CORSAllowCredentials: getEnv("CORS_ALLOW_CREDENTIALS", "true") == "true",
		CORSExposeHeaders:    getEnv("CORS_EXPOSE_HEADERS", "Content-Length"),
		CORSMaxAge:           getEnv("CORS_MAX_AGE", "86400"), //24 Hrs.
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func splitEnv(key, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	return strings.Split(value, ",")
}
