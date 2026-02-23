package main

import (
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/lib/pq"

	"github.com/hatsubosi/buygo-api/internal/adapter/db"
	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres/model"
)

func main() {
	_ = godotenv.Load()

	dbConfig := db.Config{
		Host:     envOrDefault("DB_HOST", "localhost"),
		Port:     envOrDefault("DB_PORT", "5432"),
		User:     envOrDefault("DB_USER", "buygo"),
		Password: envOrDefault("DB_PASSWORD", "local-dev-password"),
		DBName:   envOrDefault("DB_NAME", "buygo"),
		SSLMode:  envOrDefault("DB_SSLMODE", "disable"),
	}

	database, err := db.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Auto-migrate all models
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
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	log.Println("✅ Migration done")

	// ── Users ────────────────────────────────────────────────
	// Mock token format: mock-token-<uid>
	//   admin  → role 3 (SYS_ADMIN) — use token: mock-token-admin
	//   mgr    → role 2 (CREATOR)   — use token: mock-token-mgr
	//   user1  → role 1 (USER)      — use token: mock-token-user1
	users := []model.User{
		{ID: "admin", Email: "admin@example.com", Name: "Admin User", Role: 3},
		{ID: "mgr", Email: "mgr@example.com", Name: "Manager Chan", Role: 2},
		{ID: "user1", Email: "user1@example.com", Name: "Test User", Role: 1},
	}
	for _, u := range users {

		var existing model.User
		if database.Where("id = ?", u.ID).First(&existing).Error != nil {
			if err := database.Create(&u).Error; err != nil {
				log.Fatalf("Failed to create user %s: %v", u.ID, err)
			}
			log.Printf("Created user: %s (%s)", u.Name, u.Email)
		} else {
			database.Model(&existing).Updates(map[string]any{"role": u.Role, "name": u.Name})
			log.Printf("Updated user: %s", u.Email)
		}
	}

	// ── GroupBuys ────────────────────────────────────────────
	deadline := time.Now().Add(30 * 24 * time.Hour)

	groupBuys := []model.GroupBuy{
		{
			ID:             "gb-1",
			Title:          "日本零食団購",
			Description:    "精選日本零食，包含各種限定口味。費用以到貨時匯率結算。",
			Status:         2, // Active
			CreatorID:      "mgr",
			ExchangeRate:   0.22,
			SourceCurrency: "JPY",
			RoundingMethod: 1, // ROUND_DOWN
			RoundingDigit:  0,
			Deadline:       &deadline,
			ShippingConfigs: []*model.ShippingConfig{
				{ID: uuid.New().String(), Name: "宅配到府", Type: 1, Price: 150},
				{ID: uuid.New().String(), Name: "超商取貨", Type: 2, Price: 60},
			},
		},
		{
			ID:             "gb-2",
			Title:          "韓國泡麵特賣",
			Description:    "韓國直送各大品牌泡麵，數量有限。",
			Status:         2, // Active
			CreatorID:      "mgr",
			ExchangeRate:   1,
			SourceCurrency: "TWD",
			Deadline:       &deadline,
			ShippingConfigs: []*model.ShippingConfig{
				{ID: uuid.New().String(), Name: "宅配到府", Type: 1, Price: 120},
			},
		},
	}
	for _, gb := range groupBuys {
		var existing model.GroupBuy
		if database.Where("id = ?", gb.ID).First(&existing).Error != nil {
			if err := database.Create(&gb).Error; err != nil {
				log.Fatalf("Failed to create groupBuy %s: %v", gb.ID, err)
			}
			log.Printf("Created groupBuy: %s", gb.Title)
		} else {
			log.Printf("GroupBuy already exists: %s", gb.Title)
		}
	}

	// ── Products ─────────────────────────────────────────────
	products := []model.Product{
		{
			ID:            "p-1",
			GroupBuyID:    "gb-1",
			Name:          "北海道白色戀人（白巧克力）",
			Description:   "北海道限定白巧克力夾心餅乾，12片裝",
			PriceOriginal: 1500, // JPY
			ExchangeRate:  0.22,
			PriceFinal:    330, // TWD
			MaxQuantity:   10,
			Specs: []*model.ProductSpec{
				{ID: "ps-1-1", ProductID: "p-1", Name: "白巧克力 12片"},
				{ID: "ps-1-2", ProductID: "p-1", Name: "黑巧克力 12片"},
			},
		},
		{
			ID:            "p-2",
			GroupBuyID:    "gb-1",
			Name:          "東京香蕉蛋糕",
			Description:   "東京站限定香蕉夾心蛋糕，8個裝",
			PriceOriginal: 1200,
			ExchangeRate:  0.22,
			PriceFinal:    264,
			MaxQuantity:   5,
		},
		{
			ID:            "p-3",
			GroupBuyID:    "gb-2",
			Name:          "農心辛拉麵（5包裝）",
			Description:   "韓國最受歡迎拉麵，辛辣口味",
			PriceOriginal: 0,
			ExchangeRate:  1,
			PriceFinal:    80,
			MaxQuantity:   20,
		},
		{
			ID:            "p-4",
			GroupBuyID:    "gb-2",
			Name:          "三養火辣雞麵（5包裝）",
			Description:   "挑戰辣度！韓國超辣炒麵",
			PriceOriginal: 0,
			ExchangeRate:  1,
			PriceFinal:    95,
			MaxQuantity:   20,
			Specs: []*model.ProductSpec{
				{ID: "ps-4-1", ProductID: "p-4", Name: "原味"},
				{ID: "ps-4-2", ProductID: "p-4", Name: "起司口味"},
				{ID: "ps-4-3", ProductID: "p-4", Name: "奶油口味"},
			},
		},
	}
	for _, p := range products {
		var existing model.Product
		if database.Where("id = ?", p.ID).First(&existing).Error != nil {
			if err := database.Create(&p).Error; err != nil {
				log.Fatalf("Failed to create product %s: %v", p.ID, err)
			}
			log.Printf("Created product: %s", p.Name)
		} else {
			log.Printf("Product already exists: %s", p.Name)
		}
	}

	// ── Sample Order (user1 on gb-1) ─────────────────────────
	var existingOrder model.Order
	if database.Where("id = ?", "order-1").First(&existingOrder).Error != nil {
		order := model.Order{
			ID:               "order-1",
			GroupBuyID:       "gb-1",
			UserID:           "user1",
			TotalAmount:      330,
			PaymentStatus:    1, // SUBMITTED
			ContactInfo:      "Line: @testuser",
			ShippingAddress:  "台北市大安區測試路1號",
			ShippingMethodID: "",
			Items: []*model.OrderItem{
				{
					ID:          uuid.New().String(),
					OrderID:     "order-1",
					ProductID:   "p-1",
					SpecID:      "ps-1-1",
					Quantity:    1,
					ProductName: "北海道白色戀人（白巧克力）",
					SpecName:    "白巧克力 12片",
					Price:       330,
				},
			},
		}
		if err := database.Create(&order).Error; err != nil {
			log.Fatalf("Failed to create order: %v", err)
		}
		log.Printf("Created sample order: %s", order.ID)
	} else {
		log.Println("Sample order already exists")
	}

	// ── Events ───────────────────────────────────────────────
	eventStart := time.Now().Add(7 * 24 * time.Hour)
	eventEnd := time.Now().Add(10 * 24 * time.Hour)
	regDeadline := time.Now().Add(5 * 24 * time.Hour)

	var existingEvent model.Event
	if database.Where("id = ?", "ev-1").First(&existingEvent).Error != nil {
		event := model.Event{
			ID:                   "ev-1",
			Title:                "歡迎聚餐 2026 春季",
			Description:          "春季員工聚餐，請大家準時出席！套餐選項請提早確認。",
			StartTime:            eventStart,
			EndTime:              eventEnd,
			RegistrationDeadline: regDeadline,
			Location:             "台北市信義區 ATT 4 Fun B1 餐廳",
			Status:               1, // Open
			CreatorID:            "mgr",
			PaymentMethods:       pq.StringArray{"bank_transfer", "line_pay"},
			AllowException:       true,
			Items: []*model.EventItem{
				{
					ID:              "ei-1-1",
					EventID:         "ev-1",
					Name:            "全餐（主食 + 甜點 + 飲料）",
					Price:           880,
					MinParticipants: 5,
					MaxParticipants: 30,
					AllowMultiple:   false,
				},
				{
					ID:              "ei-1-2",
					EventID:         "ev-1",
					Name:            "素食套餐",
					Price:           820,
					MinParticipants: 1,
					MaxParticipants: 10,
					AllowMultiple:   false,
				},
			},
			Discounts: []*model.DiscountRule{
				{
					ID:               uuid.New().String(),
					EventID:          "ev-1",
					MinQuantity:      10,
					MinDistinctItems: 0,
					DiscountAmount:   500,
				},
			},
		}
		if err := database.Create(&event).Error; err != nil {
			log.Fatalf("Failed to create event: %v", err)
		}
		log.Printf("Created event: %s", event.Title)
	} else {
		log.Println("Event already exists")
	}

	log.Println("✅ Seed completed successfully!")
	log.Println()
	log.Println("Test tokens (use with mock auth):")
	log.Println("  Admin (SYS_ADMIN) : mock-token-admin")
	log.Println("  Manager (CREATOR) : mock-token-mgr")
	log.Println("  User              : mock-token-user1")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
