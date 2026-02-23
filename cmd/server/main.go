package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/joho/godotenv"

	"github.com/hatsubosi/buygo-api/api/v1/buygov1connect"
	"github.com/hatsubosi/buygo-api/internal/adapter/auth"
	"github.com/hatsubosi/buygo-api/internal/adapter/db"
	"github.com/hatsubosi/buygo-api/internal/adapter/handler"
	"github.com/hatsubosi/buygo-api/internal/adapter/interceptor"
	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres"
	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/hatsubosi/buygo-api/internal/service"
)

func main() {
	// 0. Setup structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 1. Load .env
	_ = godotenv.Load()

	port := envOrDefault("PORT", "8080")
	appEnv := envOrDefault("APP_ENV", "development")

	dbPassword := os.Getenv("DB_PASSWORD")
	jwtSecret := os.Getenv("JWT_SECRET")

	if appEnv == "production" {
		if dbPassword == "" {
			slog.Error("DB_PASSWORD is required in production")
			os.Exit(1)
		}
		if len(jwtSecret) < 32 {
			slog.Error("JWT_SECRET must be at least 32 characters in production")
			os.Exit(1)
		}
	} else {
		if dbPassword == "" {
			dbPassword = "password"
		}
		if jwtSecret == "" {
			jwtSecret = "secret-key"
		}
	}

	// 1. DB Connection
	dbConfig := db.Config{
		Host:     envOrDefault("DB_HOST", "localhost"),
		Port:     envOrDefault("DB_PORT", "5432"),
		User:     envOrDefault("DB_USER", "buygo"),
		Password: dbPassword,
		DBName:   envOrDefault("DB_NAME", "buygo"),
		SSLMode:  envOrDefault("DB_SSLMODE", "disable"),
	}

	database, err := db.Connect(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// 2. Migration
	slog.Info("Running AutoMigrate...")
	if err := database.AutoMigrate(
		&model.User{},
		&model.GroupBuy{},
		&model.Product{},
		&model.ProductSpec{},
		&model.Order{},
		&model.OrderItem{},
		&model.Event{},
		&model.EventItem{},
		&model.Registration{},
		&model.RegistrationItem{},
		&model.Category{},
		&model.DiscountRule{},
		&model.PriceTemplate{},
	); err != nil {
		slog.Error("Failed to migrate database", "error", err)
		os.Exit(1)
	}

	// 3. Infrastructure (Auth)
	var tokenProvider *auth.FirebaseProvider
	firebaseCreds := os.Getenv("FIREBASE_CREDENTIALS_JSON")

	if appEnv == "production" {
		if firebaseCreds == "" {
			slog.Error("FIREBASE_CREDENTIALS_JSON is required in production")
			os.Exit(1)
		}
		tp, err := auth.NewFirebaseProvider(context.Background(), []byte(firebaseCreds))
		if err != nil {
			slog.Error("Failed to init firebase", "error", err)
			os.Exit(1)
		}
		tokenProvider = tp
	} else {
		if firebaseCreds != "" {
			tp, err := auth.NewFirebaseProvider(context.Background(), []byte(firebaseCreds))
			if err != nil {
				slog.Error("Failed to init firebase", "error", err)
				os.Exit(1)
			}
			tokenProvider = tp
		} else if os.Getenv("ENABLE_MOCK_AUTH") == "true" {
			slog.Warn("Using Firebase Mock Mode — do NOT use in production")
			tokenProvider = &auth.FirebaseProvider{MockMode: true}
		} else {
			slog.Error("Set FIREBASE_CREDENTIALS_JSON or ENABLE_MOCK_AUTH=true")
			os.Exit(1)
		}
	}
	tokenManager := auth.NewJWTGenerator(jwtSecret, "buygo", 24*time.Hour)

	// 4. Repositories
	userRepo := postgres.NewUserRepository(database)
	groupBuyRepo := postgres.NewGroupBuyRepository(database)
	eventRepo := postgres.NewEventRepository(database)

	// 5. Services
	authSvc := service.NewAuthService(userRepo, tokenProvider, tokenManager)
	groupBuySvc := service.NewGroupBuyService(groupBuyRepo)
	eventSvc := service.NewEventService(eventRepo)

	// 6. Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	groupBuyHandler := handler.NewGroupBuyHandler(groupBuySvc)
	eventHandler := handler.NewEventHandler(eventSvc)

	// 7. Interceptors
	authInterceptor := interceptor.NewAuthInterceptor(tokenManager)

	// 8. Router
	mux := http.NewServeMux()

	path, handler := buygov1connect.NewAuthServiceHandler(authHandler,
		connect.WithInterceptors(authInterceptor.NewUnaryInterceptor()))
	mux.Handle(path, handler)

	path, handler = buygov1connect.NewGroupBuyServiceHandler(groupBuyHandler,
		connect.WithInterceptors(authInterceptor.NewUnaryInterceptor()))
	mux.Handle(path, handler)

	path, handler = buygov1connect.NewEventServiceHandler(eventHandler,
		connect.WithInterceptors(authInterceptor.NewUnaryInterceptor()))
	mux.Handle(path, handler)

	// 9. Health Check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := database.DB()
		if err != nil || sqlDB.Ping() != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, writeErr := w.Write([]byte(`{"status":"unhealthy"}`)); writeErr != nil {
				log.Printf("failed to write health response: %v", writeErr)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write([]byte(`{"status":"ok"}`)); writeErr != nil {
			log.Printf("failed to write health response: %v", writeErr)
		}
	})

	// 10. Server with Graceful Shutdown
	corsOrigin := envOrDefault("CORS_ORIGIN", "http://localhost:4200")
	corsHandler := newCORS(corsOrigin)

	rl := newRateLimiterFromEnv()
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           h2c.NewHandler(rl(securityHeaders(corsHandler.Handler(mux))), &http2.Server{}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Server forced to shutdown", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Server starting", "port", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Failed to serve", "error", err)
		os.Exit(1)
	}

	sqlDB, _ := database.DB()
	if sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}
	slog.Info("Server stopped")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
