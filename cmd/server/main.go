package main

import (
	"context"
	"log"
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
	// 0. Load .env
	_ = godotenv.Load()

	port := envOrDefault("PORT", "8080")
	appEnv := envOrDefault("APP_ENV", "development")

	dbPassword := os.Getenv("DB_PASSWORD")
	jwtSecret := os.Getenv("JWT_SECRET")

	if appEnv == "production" {
		if dbPassword == "" {
			log.Fatal("DB_PASSWORD is required in production")
		}
		if len(jwtSecret) < 32 {
			log.Fatal("JWT_SECRET must be at least 32 characters in production")
		}
	} else {
		// Development defaults for convenience if not set
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
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 2. Migration
	log.Println("Running AutoMigrate...")
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
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 3. Infrastructure (Auth)
	var tokenProvider *auth.FirebaseProvider
	firebaseCreds := os.Getenv("FIREBASE_CREDENTIALS_JSON")

	if appEnv == "production" {
		if firebaseCreds == "" {
			log.Fatal("FIREBASE_CREDENTIALS_JSON is required in production")
		}
		// Init real firebase
		tp, err := auth.NewFirebaseProvider(context.Background(), []byte(firebaseCreds))
		if err != nil {
			log.Fatalf("Failed to init firebase: %v", err)
		}
		tokenProvider = tp
	} else {
		// Development
		if firebaseCreds != "" {
			tp, err := auth.NewFirebaseProvider(context.Background(), []byte(firebaseCreds))
			if err != nil {
				log.Fatalf("Failed to init firebase: %v", err)
			}
			tokenProvider = tp
		} else {
			// Mock mode
			log.Println("WARNING: Using Firebase Mock Mode")
			tokenProvider = &auth.FirebaseProvider{MockMode: true}
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
			w.Write([]byte(`{"status":"unhealthy"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// 10. Server with Graceful Shutdown
	corsOrigin := envOrDefault("CORS_ORIGIN", "http://localhost:4200")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h2c.NewHandler(newCORS(corsOrigin).Handler(mux), &http2.Server{}),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Printf("Starting server on :%s\n", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to serve: %v", err)
	}

	// Cleanup
	sqlDB, _ := database.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
	log.Println("Server stopped")
}

func newCORS(origin string) *cors {
	return &cors{origin: origin}
}

type cors struct {
	origin string
}

func (c *cors) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", c.origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Connect-Protocol-Version")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
