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
	"backend/internal/handler/ws"
	"backend/internal/pkg/cloudinary"
	"backend/internal/pkg/r2"
	"backend/internal/pkg/validator"
	"backend/internal/repository/postgres"
	"backend/internal/repository/sqlc"
	authuc "backend/internal/usecase/auth"
	convuc "backend/internal/usecase/conversation"
	frienduc "backend/internal/usecase/friendship"
	msguc "backend/internal/usecase/message"
	notifuc "backend/internal/usecase/notification"
	uploaduc "backend/internal/usecase/upload"
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
	convRepo := postgres.NewConversationRepository(dbPool)
	msgRepo := postgres.NewMessageRepository(dbPool)
	friendRepo := postgres.NewFriendshipRepository(dbPool, queries)
	notifRepo := postgres.NewNotificationRepository(dbPool, queries)

	// 4. Shared Utils
	customVal := validator.New()

	// 5. Init UseCases
	registerUC := authuc.NewRegisterUseCase(userRepo, tokenRepo, cfg, customVal)
	loginUC := authuc.NewLoginUseCase(userRepo, tokenRepo, cfg, customVal)
	refreshUC := authuc.NewRefreshTokenUseCase(tokenRepo, cfg)
	logoutUC := authuc.NewLogoutUseCase(userRepo, tokenRepo)

	// Conversation UseCases
	listConvUC := convuc.NewListConversationsUseCase(convRepo)
	getConvUC := convuc.NewGetConversationUseCase(convRepo)
	createGroupUC := convuc.NewCreateGroupUseCase(convRepo, dbPool)
	markReadUC := convuc.NewMarkReadUseCase(convRepo, msgRepo)
	kickMemberUC := convuc.NewKickMemberUseCase(convRepo, dbPool)

	// Message UseCases
	sendMsgUC := msguc.NewSendMessageUseCase(convRepo, msgRepo, dbPool)
	listMsgUC := msguc.NewListMessagesUseCase(convRepo, msgRepo)
	softDeleteMsgUC := msguc.NewSoftDeleteUseCase(msgRepo)

	// Friendship UseCases
	sendReqUC := frienduc.NewSendRequestUseCase(friendRepo, userRepo, notifRepo, dbPool)
	respondReqUC := frienduc.NewRespondRequestUseCase(dbPool, friendRepo, convRepo, notifRepo, userRepo)
	unfriendUC := frienduc.NewUnfriendUseCase(friendRepo)
	listFriendUC := frienduc.NewListUseCase(friendRepo, userRepo)

	// Notification UseCases
	listNotifUC := notifuc.NewListUseCase(notifRepo)
	markNotifReadUC := notifuc.NewMarkReadUseCase(notifRepo)

	// Upload UseCases
	cldClient := cloudinary.NewClient(cloudinary.Config{
		CloudName: cfg.Cloudinary.CloudName,
		APIKey:    cfg.Cloudinary.APIKey,
		APISecret: cfg.Cloudinary.APISecret,
	})
	r2Client, err := r2.NewClient(r2.Config{
		AccountID:       cfg.R2.AccountID,
		AccessKeyID:     cfg.R2.AccessKeyID,
		SecretAccessKey: cfg.R2.SecretAccessKey,
		BucketName:      cfg.R2.BucketName,
	})
	if err != nil {
		log.Fatalf("cannot init r2 client: %v", err)
	}
	uploadUC := uploaduc.NewUsecase(cldClient, r2Client, convRepo)

	// 6. Init WebSocket Hub

	hub := ws.NewHub(cfg.Database.ListenURL, dbPool)
	hub.Run()

	// 7. Init Handlers
	authHandler := router_http.NewAuthHandler(registerUC, loginUC, refreshUC, logoutUC)
	userHandler := router_http.NewUserHandler(userRepo)
	convHandler := router_http.NewConversationHandler(listConvUC, getConvUC, createGroupUC, markReadUC, kickMemberUC)
	msgHandler := router_http.NewMessageHandler(sendMsgUC, listMsgUC, softDeleteMsgUC)
	friendHandler := router_http.NewFriendHandler(sendReqUC, respondReqUC, unfriendUC, listFriendUC)
	notifHandler := router_http.NewNotificationHandler(listNotifUC, markNotifReadUC)
	wsHandler := ws.NewHandler(hub, cfg)
	uploadHandler := router_http.NewUploadHandler(uploadUC)

	// 8. Init Router
	r := router_http.NewRouter(cfg, authHandler, userHandler, convHandler, msgHandler, friendHandler, notifHandler, wsHandler, uploadHandler)

	// 9. Start HTTP Server with Graceful Shutdown
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

	// Shutdown HTTP server
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("Server Shutdown Error: %v", err)
	}

	// Close WebSocket Hub (closes all connections and listenConn)
	if err := hub.Close(); err != nil {
		log.Printf("Hub Close Error: %v", err)
	}

	log.Println("Server exited properly")
}
