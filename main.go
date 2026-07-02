// @title           Study Tracker API
// @version         1.0
// @description     REST API for the study tracker backend
//
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT token. Format: Bearer {token}
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"study-tracker-backend/internal/auth"
	"study-tracker-backend/internal/config"
	"study-tracker-backend/internal/database"
	"study-tracker-backend/internal/handlers"
	"study-tracker-backend/internal/repository"

	_ "study-tracker-backend/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatal(err)
	}

	if err := database.EnsureDefaultAdmin(ctx, pool); err != nil {
		log.Fatal(err)
	}

	userRepo := repository.NewUserRepository(pool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(pool)
	authService := auth.NewService(cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(userRepo, refreshTokenRepo, authService, cfg.CookieSecure)
	userHandler := handlers.NewUserHandler(userRepo, authService)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/swagger/*", httpSwagger.Handler())
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)
		r.Post("/auth/logout", authHandler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(authService.Middleware)

			r.Get("/users/me", userHandler.Me)
			r.Get("/users/{id}", userHandler.Get)
			r.Post("/users", userHandler.Create)
			r.Get("/users", userHandler.List)
			r.Put("/users/{id}", userHandler.Update)
			r.Delete("/users/{id}", userHandler.Delete)
		})
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
