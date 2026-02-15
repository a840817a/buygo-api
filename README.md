# BuyGo API

Backend service for the BuyGo application, built with Go and ConnectRPC.

## Architecture

- **Language**: Go
- **Framework**: ConnectRPC (gRPC compatible)
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: Firebase Auth + JWT

## Project Structure

- `api/v1`: Protobuf definitions and generated code.
- `cmd/server`: Application entry point.
- `internal/adapter`: Infrastructure adapters (Handler, Repository, Auth, DB).
- `internal/domain`: Domain entities and interfaces.
- `internal/service`: Business logic.

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL
- Firebase Project credentials

### Running Locally

1.  Set up environment variables (see `.env.example` or strict env checks in `main.go`).
2.  Run the server:

    ```bash
    go run cmd/server/main.go
    ```

### Running Tests

Run all unit tests:

```bash
go test ./internal/...
```

## API Documentation

The API is defined using Protobuf. See `api/v1/*.proto` for service definitions.
