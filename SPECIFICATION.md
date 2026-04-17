# BuyGo Technical Specification

## 1. Constants & Enums

### 1.1. Common
```go
// User Roles
enum UserRole {
  ROLE_USER = 0;
  ROLE_CREATOR = 1;
  ROLE_SYS_ADMIN = 2;
}
```

### 1.2. Project Domain
```go
// Project Status
enum ProjectStatus {
  PROJECT_STATUS_DRAFT = 0;
  PROJECT_STATUS_ACTIVE = 1;
  PROJECT_STATUS_ENDED = 2;
  PROJECT_STATUS_ARCHIVED = 3;
}

// Order Item Status (Batch Workflows)
enum OrderItemStatus {
  ITEM_STATUS_UNORDERED = 0;
  ITEM_STATUS_ORDERED = 1;
  ITEM_STATUS_ARRIVED_OVERSEAS = 2;
  ITEM_STATUS_ARRIVED_DOMESTIC = 3;
  ITEM_STATUS_READY_FOR_PICKUP = 4;
  ITEM_STATUS_SENT = 5;
  ITEM_STATUS_FAILED = 6;
}

// Payment Status
enum PaymentStatus {
  PAYMENT_STATUS_UNSET = 0;
  PAYMENT_STATUS_SUBMITTED = 1;
  PAYMENT_STATUS_CONFIRMED = 2;
  PAYMENT_STATUS_REJECTED = 3;
}
```

### 1.3. Event Domain
```go
enum EventStatus {
  EVENT_STATUS_DRAFT = 1;
  EVENT_STATUS_ACTIVE = 2;
  EVENT_STATUS_ENDED = 3;
  EVENT_STATUS_ARCHIVED = 4;
}

enum RegistrationStatus {
  REGISTRATION_STATUS_PENDING = 1;
  REGISTRATION_STATUS_CONFIRMED = 2;
  REGISTRATION_STATUS_CANCELLED = 3;
}
```

## 2. Data Models (Domain Entities)

### 2.1. Pricing Models
**Rounding Config**
- `Method`: `FLOOR`, `CEIL`, `ROUND`.
- `Digit`: Integer.
    - `0`: Round to ones (123.4 -> 123).
    - `1`: Round to tens (123 -> 120/130).
    - `-1`: Round to tenths (123.45 -> 123.5).

**Formula**:
`FinalPrice = Round(OriginalPrice * ExchangeRate, Config)`

### 2.2. Product Batch Logic (FIFO)
**Scenario**: Manager inputs "5 items arrived".
**Algorithm**:
1.  **Input**: `SpecID`, `TargetStatus`, `Count`.
2.  **Select**: Query DB for `OrderItem` where `SpecId == input.SpecID` AND `Status < input.TargetStatus`.
3.  **Sort**: Ascending by `CreatedAt` (Oldest First).
4.  **Limit**: Take top `Count`.
5.  **Update**: Set `Status = TargetStatus` for these items.

## 3. API Definitions (Proto Summary)

### 3.1. Project Service
**RPCs**:
- `CreateProject(CreateProjectRequest) returns (CreateProjectResponse)`
- `ListProjects(ListProjectsRequest) returns (ListProjectsResponse)`
- `GetProject(GetProjectRequest) returns (GetProjectResponse)`
- `UpdateProject(UpdateProjectRequest) returns (UpdateProjectResponse)`
- `AddProduct(AddProductRequest) returns (AddProductResponse)`
- `CreateOrder(CreateOrderRequest) returns (CreateOrderResponse)`
- `CancelOrder(CancelOrderRequest) returns (CancelOrderResponse)`
- `GetMyOrders(GetMyOrdersRequest) returns (GetMyOrdersResponse)`
- `ListProjectOrders(ListProjectOrdersRequest) returns (ListProjectOrdersResponse)` (Manager Only)
- `ConfirmPayment(ConfirmPaymentRequest) returns (ConfirmPaymentResponse)` (Manager Only)
- `BatchUpdateStatus(BatchUpdateStatusRequest) returns (BatchUpdateStatusResponse)` (Manager Only)

**Key Structures**:
```protobuf
message CreateOrderRequest {
  string project_id = 1;
  repeated CreateOrderItem items = 2;
  string contact_info = 3;
  string shipping_address = 4;
}

message BatchUpdateStatusRequest {
  string project_id = 1;
  string spec_id = 2;
  OrderItemStatus target_status = 3;
  int32 count = 4; // Number of items to progress
}
```

### 3.2. Event Service
**RPCs**:
- `CreateEvent(CreateEventRequest) returns (CreateEventResponse)`
- `ListEvents(ListEventsRequest) returns (ListEventsResponse)`
- `GetEvent(GetEventRequest) returns (GetEventResponse)`
- `RegisterEvent(RegisterEventRequest) returns (RegisterEventResponse)`
- `CancelRegistration(CancelRegistrationRequest) returns (CancelRegistrationResponse)`
- `GetMyRegistrations(GetMyRegistrationsRequest) returns (GetMyRegistrationsResponse)`

**Key Structures**:
```protobuf
message CreateEventRequest {
  string title = 1;
  string description = 2;
  google.protobuf.Timestamp start_time = 3;
  google.protobuf.Timestamp end_time = 4;
}

message RegisterEventRequest {
  string event_id = 1;
  repeated RegisterItem items = 2;
  string contact_info = 3;
  string notes = 4;
}
```

### 3.2. Order Modifications
- Users can modify orders *only if* `Status == UNORDERED`.
- Once `Status >= ORDERED`, modification requires Admin approval (or is blocked).

## 4. Workflows

### 4.1. Order Lifecycle
1.  **User** places Order (`Status=UNORDERED`).
2.  **Manager** collects confirmed orders.
3.  **Manager** uses "Batch Update: Ordered" for total count (e.g., 50).
    - System updates 50 oldest `UNORDERED` items -> `ORDERED`.
4.  **Manager** updates "Arrived Overseas".
    - System updates `ORDERED` items -> `ARRIVED_OVERSEAS`.
    - Visualization: "30/50 Arrived".

### 4.2. Payment Verification
1.  **User** submits Payment Proof (Method: Bank Transfer, Last5: 12345, Time: ...).
    - `PaymentStatus` -> `SUBMITTED`.
2.  **Manager** sees list of `SUBMITTED`.
3.  **Manager** clicks "Confirm".
    - `PaymentStatus` -> `CONFIRMED`.
