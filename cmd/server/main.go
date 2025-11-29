package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/yourusername/golf_messenger/internal/config"
	"github.com/yourusername/golf_messenger/internal/database"
	"github.com/yourusername/golf_messenger/internal/handler"
	"github.com/yourusername/golf_messenger/internal/logger"
	"github.com/yourusername/golf_messenger/internal/repository"
	"github.com/yourusername/golf_messenger/internal/router"
	"github.com/yourusername/golf_messenger/internal/service"
	"github.com/yourusername/golf_messenger/pkg/storage"
	"go.uber.org/zap"
)

// @title Golf Messenger API
// @version 1.0
// @description Golf tee time reservation and messaging platform API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.golfmessenger.com/support
// @contact.email support@golfmessenger.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Golf Messenger API server",
		zap.String("version", "1.0"),
		zap.String("port", cfg.Server.Port),
	)

	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	log.Info("Database connected successfully")

	s3Client, err := storage.NewS3Client(&cfg.AWS)
	if err != nil {
		log.Fatal("Failed to initialize S3 client", zap.Error(err))
	}

	log.Info("S3 client initialized successfully")

	userRepo := repository.NewUserRepository(db.DB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db.DB)
	ttrRepo := repository.NewTTRRepository(db.DB)
	invitationRepo := repository.NewInvitationRepository(db.DB)

	notificationService := service.NewNotificationService(log)

	authService := service.NewAuthService(
		userRepo,
		refreshTokenRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)
	userService := service.NewUserService(userRepo, s3Client)
	ttrService := service.NewTTRService(ttrRepo, userRepo, log)
	invitationService := service.NewInvitationService(invitationRepo, ttrRepo, userRepo, notificationService, log)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	ttrHandler := handler.NewTTRHandler(ttrService)
	invitationHandler := handler.NewInvitationHandler(invitationService)

	rt := router.NewRouter(
		authHandler,
		userHandler,
		ttrHandler,
		invitationHandler,
		log,
		cfg.JWT.Secret,
		cfg.CORS.AllowedOrigins,
	)

	httpHandler := rt.SetupRoutes()

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      httpHandler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Info("Server starting", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server shutdown complete")
}
