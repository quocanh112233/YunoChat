package http

import (
	"backend/internal/config"
	"backend/internal/handler/http/middleware"
	"backend/internal/handler/ws"

	chi_v5 "github.com/go-chi/chi/v5"
	chi_mdw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	cfg *config.Config,
	authH *AuthHandler,
	userH *UserHandler,
	convH *ConversationHandler,
	msgH *MessageHandler,
	friendH *FriendHandler,
	notifH *NotificationHandler,
	wsHandler *ws.Handler,
	uploadH *UploadHandler,
) *chi_v5.Mux {
	r := chi_v5.NewRouter()

	// 1. Basic Middleware Stack
	r.Use(chi_mdw.RequestID)
	r.Use(chi_mdw.RealIP)
	r.Use(chi_mdw.Logger)
	r.Use(chi_mdw.Recoverer)

	// 2. Custom Middlewares
	r.Use(middleware.CORS(cfg.Server.AllowedOrigins))
	r.Use(middleware.RateLimit(30, 100)) // 30 req/s, burst 100

	// 3. API Routes setup
	r.Route("/v1", func(r chi_v5.Router) {
		// Public Auth Routes
		r.Route("/auth", func(r chi_v5.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.Refresh)

			// Logout requires Auth (UserID)
			r.Group(func(r chi_v5.Router) {
				r.Use(middleware.RequireAuth(cfg.JWT.AccessSecret))
				r.Post("/logout", authH.Logout)
			})
		})

		// WebSocket endpoint - public but requires valid token in query param
		r.Get("/ws", wsHandler.HandleUpgrade)

		// Protected Routes
		r.Group(func(r chi_v5.Router) {
			r.Use(middleware.RequireAuth(cfg.JWT.AccessSecret))

			// Users endpoints
			r.Route("/users", func(r chi_v5.Router) {
				r.Get("/me", userH.GetMe)
				r.Patch("/me", userH.UpdateMe)
			})

			// Conversation endpoints
			convH.RegisterRoutes(r)

			// Message endpoints
			msgH.RegisterRoutes(r)

			// Friend endpoints
			friendH.RegisterRoutes(r)

			// Notification endpoints
			notifH.RegisterRoutes(r)

			// Upload endpoints
			uploadH.RegisterRoutes(r)
		})
	})

	return r
}
