package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

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
	port := envOrDefault("PORT", "8080")

	// 1. DB Connection
	dbConfig := db.Config{
		Host:     envOrDefault("DB_HOST", "localhost"),
		Port:     envOrDefault("DB_PORT", "5432"),
		User:     envOrDefault("DB_USER", "buygo"),
		Password: envOrDefault("DB_PASSWORD", "password"),
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
		&model.Project{},
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
	jwtSecret := envOrDefault("JWT_SECRET", "secret-key")
	tokenProvider := &auth.FirebaseProvider{}
	tokenManager := auth.NewJWTGenerator(jwtSecret, "buygo", 24*time.Hour)

	// 4. Repositories
	userRepo := postgres.NewUserRepository(database)
	projectRepo := postgres.NewProjectRepository(database)
	eventRepo := postgres.NewEventRepository(database)

	// 5. Services
	authSvc := service.NewAuthService(userRepo, tokenProvider, tokenManager)
	projectSvc := service.NewGroupBuyService(projectRepo)
	eventSvc := service.NewEventService(eventRepo)

	// 6. Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	projectHandler := handler.NewGroupBuyHandler(projectSvc)
	eventHandler := handler.NewEventHandler(eventSvc)

	// 7. Interceptors
	authInterceptor := interceptor.NewAuthInterceptor(tokenManager)

	// 8. Router
	mux := http.NewServeMux()

	path, handler := buygov1connect.NewAuthServiceHandler(authHandler,
		connect.WithInterceptors(authInterceptor.NewUnaryInterceptor()))
	mux.Handle(path, handler)

	path, handler = buygov1connect.NewGroupBuyServiceHandler(projectHandler,
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
