package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/config"
	router_http "backend/internal/handler/http"
	"backend/internal/pkg/validator"
	"backend/internal/repository/postgres"
	"backend/internal/repository/sqlc"
	authuc "backend/internal/usecase/auth"
)

func main() {
	// 1. Load config
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	// 2. Init DB Connection Pool
	ctx := context.Background()
	dbPool, err := postgres.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}
	defer dbPool.Close()

	// 3. Init Repositories
	queries := sqlc.New(dbPool)
	userRepo := postgres.NewUserRepository(queries)
	tokenRepo := postgres.NewRefreshTokenRepository(queries)

	// 4. Shared Utils
	customVal := validator.New()

	// 5. Init UseCases
	registerUC := authuc.NewRegisterUseCase(userRepo, tokenRepo, cfg, customVal)
	loginUC := authuc.NewLoginUseCase(userRepo, tokenRepo, cfg, customVal)
	refreshUC := authuc.NewRefreshTokenUseCase(tokenRepo, cfg)
	logoutUC := authuc.NewLogoutUseCase(userRepo, tokenRepo)

	// 6. Init Handlers
	authHandler := router_http.NewAuthHandler(registerUC, loginUC, refreshUC, logoutUC)
	userHandler := router_http.NewUserHandler(userRepo)

	// 7. Init Router
	r := router_http.NewRouter(cfg, authHandler, userHandler)

	// 8. Start HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	// Wait for OS signal (CTRL-C or docker stop)
	<-quit
	log.Println("Shutting down server...")

	// Create 10s timeout ctx for shutting down requests nicely
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}

	log.Println("Server exited properly")
}
