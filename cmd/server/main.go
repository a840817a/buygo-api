package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/joho/godotenv"

	"github.com/buygo/buygo-api/api/v1/buygov1connect"
	"github.com/buygo/buygo-api/internal/adapter/auth"
	"github.com/buygo/buygo-api/internal/adapter/db"
	"github.com/buygo/buygo-api/internal/adapter/handler"
	"github.com/buygo/buygo-api/internal/adapter/interceptor"
	"github.com/buygo/buygo-api/internal/adapter/repository/postgres"
	"github.com/buygo/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/buygo/buygo-api/internal/service"
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

	corsOrigin := envOrDefault("CORS_ORIGIN", "http://localhost:4200")

	fmt.Printf("Starting server on :%s\n", port)
	err = http.ListenAndServe(
		":"+port,
		h2c.NewHandler(newCORS(corsOrigin).Handler(mux), &http2.Server{}),
	)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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
