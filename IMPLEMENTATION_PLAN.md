# BuyGo Backend Implementation Plan (buygo-api)

## Goal
Implement the backend for the BuyGo system using Golang, following Clean Architecture and Domain-Driven Design (DDD).

## Architecture
**Style**: Clean Architecture (Hexagonal)
**Communication**: gRPC (ConnectRPC)

### Directory Structure
```text
buygo-api/
├── cmd/
│   └── server/          # Main application entry point
├── api/                 # Protobuf definitions (gRPC)
│   └── v1/              # Versioning
├── internal/
│   ├── domain/          # Entities & Business Logic (Pure Go)
│   │   ├── user/
│   │   ├── project/
│   │   └── event/
│   ├── port/            # Interfaces (Input/Output Ports)
│   │   ├── repository/  # DB Interfaces
│   │   └── service/     # Service Interfaces
│   ├── app/             # Application Services (Use Cases)
│   └── adapter/         # Infrastructure Implementations
│       ├── postgres/    # Database Adapter
│       ├── grpc/        # gRPC Handlers
│       └── auth/        # Auth Provider (Firebase)
```

## Features & Domains

### 1. User Domain (Auth)
- **Roles**: `USER`, `CREATOR`, `SYS_ADMIN`.
- **Features**: 
    - Verify 3rd party tokens (Firebase).
    - Manage Profiles and Roles.

### 2. Project Domain (Group Buying)
- **Entities**:
    - `Project`: Status (`DRAFT`...`ARCHIVED`), PaymentMethods, ShippingOptions.
    - `Product`: Batch Logic, Rounding Config (`CEIL/FLOOR`, `Digit`).
    - `OrderItem`: Status (`UNORDERED` -> `ORDERED` -> ... -> `SENT`).
    - `PaymentRecord`: Status (`SUBMITTED` -> `CONFIRMED`).
- **Core Logic**:
    - **FIFO Batch Update**: Update oldest `UNORDERED` items first when Manager confirms order with supplier.

### 3. Event Domain (Activity)
- **Entities**: `Event`, `EventItem` (Limits, Discounts).
- **Features**: Registration with item limits.

## Next Steps
1.  **Specification**: Refer to `SPECIFICATION.md` for full Protobuf definitions and Logic.
2.  **Codegen**: Update `.proto` files and run buf generate.
3.  **Core**: Implement `User`, `Project`, `Order` domains with new Enums/Structs.
4.  **Service**: Implement `BatchUpdate` logic.
