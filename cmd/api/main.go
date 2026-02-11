// cmd/api/main.go
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
	"github.com/go-chi/cors"
	"github.com/jd7008911/aogeri-api/internal/auth"
	"github.com/jd7008911/aogeri-api/internal/config"
	"github.com/jd7008911/aogeri-api/internal/db"
	"github.com/jd7008911/aogeri-api/internal/handlers"
	"github.com/jd7008911/aogeri-api/internal/services"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database
	database, err := db.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Run migrations
	if err := runMigrations(&cfg.Database); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Initialize services
	// Wrap redis client with store adapter
	redisStore := auth.NewRedisStore(redisClient)
	authService := auth.NewAuthService(database.Queries, cfg, redisStore)

	stakingService := services.NewStakingService(database.Queries, authService)
	dashboardService := services.NewDashboardService(database.Queries, authService)
	governanceService := services.NewGovernanceService(database.Queries, authService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, database.Queries)
	stakeHandler := handlers.NewStakeHandler(database.Queries, stakingService, authService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	governanceHandler := handlers.NewGovernanceHandler(database.Queries, governanceService)
	assetHandler := handlers.NewAssetsHandler()

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes
		authHandler.RegisterRoutes(r)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authService.AuthMiddleware)

			stakeHandler.RegisterRoutes(r)
			dashboardHandler.RegisterRoutes(r)
			governanceHandler.RegisterRoutes(r)
			assetHandler.RegisterRoutes(r)
		})
	})

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited properly")
}

func runMigrations(cfg *config.DatabaseConfig) error {
	// Use goose to run migrations
	// This is a placeholder - implement actual goose migration logic
	return nil
}
