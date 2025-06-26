package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"warehouse-store/config"
	"warehouse-store/middlewares"
	"warehouse-store/models"
	"warehouse-store/routers"
	"warehouse-store/utils"
)

func main() {
	utils.InitLogger()
	defer utils.Logger.Sync()
	cfg := config.LoadConfig()

	// Connect to PostgreSQL
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	var db *gorm.DB
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("Connected to PostgreSQL")
			break
		}
		log.Printf("Failed to connect to PostgreSQL (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(5 * time.Second) // Wait before retrying
	}
	if err != nil {
		log.Fatalf("Fatal: Could not connect to PostgreSQL after multiple retries: %v", err)
	}

	// Auto-migrate database schema
	err = db.AutoMigrate(&models.User{}, &models.Project{}, &models.Category{}, &models.Item{}, &models.TransactionBorrow{}, &models.TransactionReturn{}, &models.DamageReport{}, &models.AuditLog{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// Connect to Redis (optional for this example, but good practice for caching)
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   0, // use default DB
	})

	_, err = rdb.Ping(rdb.Context()).Result()
	if err != nil {
		log.Printf("Could not connect to Redis: %v (Redis caching will not be used)", err)
	} else {
		log.Println("Connected to Redis")
	}

	// Setup Gin router
	r := routers.SetupRouter(db)

	// Configure CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowMethods:     cfg.CORSAllowMethods,
		AllowHeaders:     cfg.CORSAllowHeaders,
		ExposeHeaders:    strings.Split(cfg.CORSExposeHeaders, ","),
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAge:           parseDuration(cfg.CORSMaxAge),
	}))

	r.Use(middlewares.AuditLogger(db))

	// Run the server
	log.Printf("Server listening on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func parseDuration(seconds string) time.Duration {
	duration, err := time.ParseDuration(seconds + "s")
	if err != nil {
		return 24 * time.Hour 
	}
	return duration
}
