# BuyGo API

Backend service for BuyGo — a group buying and event registration management platform. Built with Go and ConnectRPC, following Clean Architecture.

## Features

- **Group Buy Management** — Create and manage group buying projects with product catalogs, multi-currency pricing, and order lifecycle tracking
- **Event Registration** — Event creation with item selection, discount rules, and registration management
- **FIFO Batch Fulfillment** — Batch update order item statuses using oldest-first (FIFO) processing
- **Multi-Currency Pricing** — Price templates with configurable exchange rates and rounding strategies (floor/ceil/round)
- **Role-Based Access Control** — User, Creator, and System Admin roles with Firebase Auth + JWT

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.25 |
| RPC Framework | ConnectRPC (gRPC-compatible) |
| Database | PostgreSQL 18 |
| ORM | GORM |
| Authentication | Firebase Auth (token verification) + JWT (backend sessions) |
| Protobuf | buf (linting, code generation) |
| Testing | testify, testcontainers |

## Architecture

The project follows **Clean Architecture** (Hexagonal) with strict dependency rules — inner layers never depend on outer layers.

```
buygo-api/
├── api/v1/                    # Protobuf definitions & generated code
│   ├── *.proto                # Service definitions (source of truth)
│   └── buygov1connect/        # Generated ConnectRPC handlers
├── cmd/
│   └── server/                # Application entry point & middleware
├── internal/
│   ├── domain/                # Entities & business rules (pure Go, no dependencies)
│   │   ├── auth/
│   │   ├── user/
│   │   ├── groupbuy/
│   │   └── event/
│   ├── service/               # Application services (use cases)
│   └── adapter/               # Infrastructure implementations
│       ├── handler/           # ConnectRPC request handlers
│       ├── interceptor/       # Auth interceptors
│       ├── repository/
│       │   ├── postgres/      # PostgreSQL repository (GORM)
│       │   └── memory/        # In-memory repository (testing)
│       ├── auth/              # Firebase auth provider
│       └── db/                # Database connection
├── scripts/                   # CI/CD helper scripts
├── Dockerfile                 # Multi-stage build (Go → Alpine)
└── docker-compose.yml         # Full-stack local setup
```

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL 18 (or use Docker Compose)
- Firebase project credentials (optional — mock mode available for development)

### Environment Variables

Copy `.env.example` and configure:

```bash
cp .env.example .env
```

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | `development` or `production` | `development` |
| `JWT_SECRET` | JWT signing secret (32+ chars in production) | — |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `buygo` |
| `DB_PASSWORD` | Database password | — |
| `DB_NAME` | Database name | `buygo` |
| `DB_SSLMODE` | SSL mode (`disable` / `require`) | `disable` |
| `CORS_ORIGIN` | Allowed frontend origin | `http://localhost:4200` |
| `ENABLE_MOCK_AUTH` | Enable Firebase mock mode (dev only) | `true` |
| `RATE_LIMIT_RPS` | Rate limit (requests/sec) | `100` |
| `RATE_LIMIT_BURST` | Rate limit burst | `200` |

### Running Locally

**Direct execution:**

```bash
go run cmd/server/main.go
```

**Docker Compose** (API + PostgreSQL + Web):

```bash
docker-compose up
```

This starts:
- `buygo-db` — PostgreSQL 18 on port 5432
- `buygo-api` — API server on port 8080
- `buygo-web` — Frontend on port 4200

## API Documentation

The API is defined in Protocol Buffers (`api/v1/*.proto`). Full generated documentation is available at [`docs/api.md`](docs/api.md).

To regenerate after modifying `.proto` files:

```bash
go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
buf generate --template buf.gen.doc.yaml
```

### Overview

Three main services:

### AuthService (`auth.proto`)

| RPC | Description | Access |
|-----|-------------|--------|
| `Login` | Exchange Firebase token for JWT | Public |
| `GetMe` | Get current user profile | Authenticated |
| `ListUsers` | List all users | Admin |
| `UpdateUserRole` | Change user role | Admin |
| `ListAssignableManagers` | List manager candidates | Creator / Admin |

### GroupBuyService (`groupbuy.proto`)

| RPC | Description | Access |
|-----|-------------|--------|
| `CreateGroupBuy` | Create group buy project | Creator / Admin |
| `ListGroupBuys` | List active group buys | Public |
| `GetGroupBuy` | Get group buy with products | Public |
| `UpdateGroupBuy` | Update group buy details | Creator / Manager |
| `AddProduct` | Add product with specs | Creator / Manager |
| `CreateOrder` | Place an order | Authenticated |
| `UpdateOrder` | Modify order (only if unordered) | Order Owner |
| `CancelOrder` | Cancel order (only if unordered) | Order Owner |
| `UpdatePaymentInfo` | Submit payment proof | Order Owner |
| `GetMyOrders` / `GetMyGroupBuyOrder` | View own orders | Authenticated |
| `ListGroupBuyOrders` | View all orders for a group buy | Manager |
| `ConfirmPayment` | Approve/reject payment | Manager |
| `BatchUpdateStatus` | FIFO batch status progression | Manager |
| `CreateCategory` | Create category templates | Admin |
| `ListCategories` | List category templates | Creator / Admin |
| `CreatePriceTemplate` / `UpdatePriceTemplate` / `DeletePriceTemplate` | Manage pricing templates | Admin |
| `ListPriceTemplates` / `GetPriceTemplate` | Read pricing templates | Creator / Admin |

### EventService (`event.proto`)

| RPC | Description | Access |
|-----|-------------|--------|
| `CreateEvent` | Create event with items & discounts | Creator / Admin |
| `ListEvents` / `GetEvent` | Browse events | Public |
| `UpdateEvent` / `UpdateEventStatus` | Manage event | Creator / Manager |
| `RegisterEvent` | Register for event | Authenticated |
| `UpdateRegistration` | Modify registration | Registrant |
| `CancelRegistration` | Cancel registration | Registrant |
| `UpdateRegistrationStatus` | Approve/reject registration | Manager |
| `GetMyRegistrations` | View own registrations | Authenticated |
| `ListEventRegistrations` | View all registrations | Manager |

## Core Business Logic

### FIFO Batch Status Update

When a manager updates item statuses (e.g., "5 items arrived"), the system processes orders oldest-first:

1. Query items matching `SpecID` with `Status < TargetStatus`
2. Sort by `CreatedAt` ascending (oldest first)
3. Take top `Count` items
4. Update their status to `TargetStatus`

### Multi-Currency Pricing

```
FinalPrice = Round(OriginalPrice × ExchangeRate, RoundingConfig)
```

Rounding methods: `FLOOR`, `CEIL`, `ROUND` with configurable digit precision.

### Key Enums

| Domain | Enum | Values |
|--------|------|--------|
| GroupBuy | Status | Draft → Active → Ended → Archived |
| Order Item | Status | Unordered → Ordered → Arrived Overseas → Arrived Domestic → Ready for Pickup → Sent / Failed |
| Payment | Status | Unset → Submitted → Confirmed / Rejected |
| Event | Status | Draft → Active → Ended → Archived |
| Registration | Status | Pending → Confirmed / Cancelled |

## Testing

Run all tests:

```bash
go test ./...
```

Run with race detection and coverage:

```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

- **Coverage gate**: 70% minimum (enforced in CI)
- **Integration tests**: Use testcontainers for real PostgreSQL
- **Race detection**: Enabled in CI with `-race` flag

## CI/CD

GitHub Actions pipeline (`.github/workflows/ci.yml`) runs on push/PR to `main` and `dev`:

1. **test** — `go mod verify` → `gofmt` check → tests with race detection & coverage → coverage gate (70%) → gosec & govulncheck → build verification
2. **buf-lint** — Protobuf linting (STANDARD rules) + breaking change detection
3. **docker** — Docker image build test (depends on test passing)
4. **gitleaks** — Secret scanning
