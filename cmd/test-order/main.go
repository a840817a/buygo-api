package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"connectrpc.com/connect"
	buygov1 "github.com/buygo/buygo-api/api/v1"
	"github.com/buygo/buygo-api/api/v1/buygov1connect"
	"github.com/buygo/buygo-api/internal/adapter/auth"
	"github.com/buygo/buygo-api/internal/domain/user"
)

func main() {
	// 1. Create Token for 'dev' user
	tokenGen := auth.NewJWTGenerator("secret-key", "buygo", 24*time.Hour)
	u := &user.User{
		ID:   "dev",
		Name: "Dev Guest",
		Role: user.UserRoleSysAdmin,
	}
	token, err := tokenGen.GenerateToken(u)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	// 2. Setup Client
	client := buygov1connect.NewGroupBuyServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
	)

	// 3. Prepare Request
	// Need a valid ProjectID and ProductID/SpecID from seed
	// Project: Japan Snacks (seeded with specific UUID? No, UUID was random in seed)
	// Query DB to find IDs first? Or just hardcode if I know them?
	// Seed script used uuid.New().String(), so they are random.
	// I need to fetch them.

	// Let's use ListProjects first
	req := connect.NewRequest(&buygov1.ListGroupBuysRequest{})
	req.Header().Set("Authorization", "Bearer "+token)

	res, err := client.ListGroupBuys(context.Background(), req)
	if err != nil {
		log.Fatalf("ListProjects failed: %v", err)
	}
	if len(res.Msg.GroupBuys) == 0 {
		log.Fatalf("No projects found. Did seed run?")
	}

	var groupBuy *buygov1.GroupBuy
	for _, p := range res.Msg.GroupBuys {
		if p.Title == "Japan Snacks Group Buy" {
			groupBuy = p
			break
		}
	}
	if groupBuy == nil {
		log.Fatalf("GroupBuy 'Japan Snacks Group Buy' not found")
	}
	log.Printf("Using GroupBuy: %s (%s)", groupBuy.Title, groupBuy.Id)

	// Load Details to get Products
	detailReq := connect.NewRequest(&buygov1.GetGroupBuyRequest{GroupBuyId: groupBuy.Id})
	detailReq.Header().Set("Authorization", "Bearer "+token)
	detailRes, err := client.GetGroupBuy(context.Background(), detailReq)
	if err != nil {
		log.Fatalf("GetProject failed: %v", err)
	}

	if len(detailRes.Msg.Products) == 0 {
		log.Fatalf("No products found in project")
	}
	product := detailRes.Msg.Products[0]
	specId := ""
	if len(product.Specs) > 0 {
		specId = product.Specs[0].Id
	}
	log.Printf("Using Product: %s (%s) Spec: %s", product.Name, product.Id, specId)

	// 4. Create Order
	createReq := connect.NewRequest(&buygov1.CreateOrderRequest{
		GroupBuyId: groupBuy.Id,
		Items: []*buygov1.CreateOrderItem{
			{
				ProductId: product.Id,
				SpecId:    specId,
				Quantity:  2,
			},
		},
		ContactInfo:     "Test Go Client",
		ShippingAddress: "123 Go St",
	})
	createReq.Header().Set("Authorization", "Bearer "+token)

	log.Println("Sending CreateOrder...")
	createRes, err := client.CreateOrder(context.Background(), createReq)
	if err != nil {
		log.Fatalf("CreateOrder failed: %v", err)
	}

	log.Printf("Order Created Successfully! ID: %s", createRes.Msg.OrderId)
}
