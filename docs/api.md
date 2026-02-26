# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [api/v1/auth.proto](#api_v1_auth-proto)
    - [GetMeRequest](#buygo-v1-GetMeRequest)
    - [GetMeResponse](#buygo-v1-GetMeResponse)
    - [ListAssignableManagersRequest](#buygo-v1-ListAssignableManagersRequest)
    - [ListAssignableManagersResponse](#buygo-v1-ListAssignableManagersResponse)
    - [ListUsersRequest](#buygo-v1-ListUsersRequest)
    - [ListUsersResponse](#buygo-v1-ListUsersResponse)
    - [LoginRequest](#buygo-v1-LoginRequest)
    - [LoginResponse](#buygo-v1-LoginResponse)
    - [UpdateUserRoleRequest](#buygo-v1-UpdateUserRoleRequest)
    - [UpdateUserRoleResponse](#buygo-v1-UpdateUserRoleResponse)
    - [User](#buygo-v1-User)
  
    - [UserRole](#buygo-v1-UserRole)
  
    - [AuthService](#buygo-v1-AuthService)
  
- [api/v1/groupbuy.proto](#api_v1_groupbuy-proto)
    - [AddProductRequest](#buygo-v1-AddProductRequest)
    - [AddProductResponse](#buygo-v1-AddProductResponse)
    - [BatchUpdateStatusRequest](#buygo-v1-BatchUpdateStatusRequest)
    - [BatchUpdateStatusResponse](#buygo-v1-BatchUpdateStatusResponse)
    - [CancelOrderRequest](#buygo-v1-CancelOrderRequest)
    - [CancelOrderResponse](#buygo-v1-CancelOrderResponse)
    - [Category](#buygo-v1-Category)
    - [ConfirmPaymentRequest](#buygo-v1-ConfirmPaymentRequest)
    - [ConfirmPaymentResponse](#buygo-v1-ConfirmPaymentResponse)
    - [CreateCategoryRequest](#buygo-v1-CreateCategoryRequest)
    - [CreateCategoryResponse](#buygo-v1-CreateCategoryResponse)
    - [CreateGroupBuyRequest](#buygo-v1-CreateGroupBuyRequest)
    - [CreateGroupBuyResponse](#buygo-v1-CreateGroupBuyResponse)
    - [CreateOrderItem](#buygo-v1-CreateOrderItem)
    - [CreateOrderRequest](#buygo-v1-CreateOrderRequest)
    - [CreateOrderResponse](#buygo-v1-CreateOrderResponse)
    - [CreatePriceTemplateRequest](#buygo-v1-CreatePriceTemplateRequest)
    - [CreatePriceTemplateResponse](#buygo-v1-CreatePriceTemplateResponse)
    - [DeletePriceTemplateRequest](#buygo-v1-DeletePriceTemplateRequest)
    - [DeletePriceTemplateResponse](#buygo-v1-DeletePriceTemplateResponse)
    - [GetGroupBuyRequest](#buygo-v1-GetGroupBuyRequest)
    - [GetGroupBuyResponse](#buygo-v1-GetGroupBuyResponse)
    - [GetMyGroupBuyOrderRequest](#buygo-v1-GetMyGroupBuyOrderRequest)
    - [GetMyGroupBuyOrderResponse](#buygo-v1-GetMyGroupBuyOrderResponse)
    - [GetMyOrdersRequest](#buygo-v1-GetMyOrdersRequest)
    - [GetMyOrdersResponse](#buygo-v1-GetMyOrdersResponse)
    - [GetPriceTemplateRequest](#buygo-v1-GetPriceTemplateRequest)
    - [GetPriceTemplateResponse](#buygo-v1-GetPriceTemplateResponse)
    - [GroupBuy](#buygo-v1-GroupBuy)
    - [ListCategoriesRequest](#buygo-v1-ListCategoriesRequest)
    - [ListCategoriesResponse](#buygo-v1-ListCategoriesResponse)
    - [ListGroupBuyOrdersRequest](#buygo-v1-ListGroupBuyOrdersRequest)
    - [ListGroupBuyOrdersResponse](#buygo-v1-ListGroupBuyOrdersResponse)
    - [ListGroupBuysRequest](#buygo-v1-ListGroupBuysRequest)
    - [ListGroupBuysResponse](#buygo-v1-ListGroupBuysResponse)
    - [ListManagerGroupBuysRequest](#buygo-v1-ListManagerGroupBuysRequest)
    - [ListManagerGroupBuysResponse](#buygo-v1-ListManagerGroupBuysResponse)
    - [ListPriceTemplatesRequest](#buygo-v1-ListPriceTemplatesRequest)
    - [ListPriceTemplatesResponse](#buygo-v1-ListPriceTemplatesResponse)
    - [Order](#buygo-v1-Order)
    - [OrderItem](#buygo-v1-OrderItem)
    - [PaymentInfo](#buygo-v1-PaymentInfo)
    - [PriceTemplate](#buygo-v1-PriceTemplate)
    - [Product](#buygo-v1-Product)
    - [ProductSpec](#buygo-v1-ProductSpec)
    - [RoundingConfig](#buygo-v1-RoundingConfig)
    - [ShippingConfig](#buygo-v1-ShippingConfig)
    - [UpdateGroupBuyRequest](#buygo-v1-UpdateGroupBuyRequest)
    - [UpdateGroupBuyResponse](#buygo-v1-UpdateGroupBuyResponse)
    - [UpdateOrderRequest](#buygo-v1-UpdateOrderRequest)
    - [UpdateOrderResponse](#buygo-v1-UpdateOrderResponse)
    - [UpdatePaymentInfoRequest](#buygo-v1-UpdatePaymentInfoRequest)
    - [UpdatePaymentInfoResponse](#buygo-v1-UpdatePaymentInfoResponse)
    - [UpdatePriceTemplateRequest](#buygo-v1-UpdatePriceTemplateRequest)
    - [UpdatePriceTemplateResponse](#buygo-v1-UpdatePriceTemplateResponse)
  
    - [GroupBuyStatus](#buygo-v1-GroupBuyStatus)
    - [OrderItemStatus](#buygo-v1-OrderItemStatus)
    - [PaymentStatus](#buygo-v1-PaymentStatus)
    - [RoundingMethod](#buygo-v1-RoundingMethod)
    - [ShippingType](#buygo-v1-ShippingType)
  
    - [GroupBuyService](#buygo-v1-GroupBuyService)
  
- [api/v1/event.proto](#api_v1_event-proto)
    - [CancelRegistrationRequest](#buygo-v1-CancelRegistrationRequest)
    - [CancelRegistrationResponse](#buygo-v1-CancelRegistrationResponse)
    - [CreateEventRequest](#buygo-v1-CreateEventRequest)
    - [CreateEventResponse](#buygo-v1-CreateEventResponse)
    - [DiscountRule](#buygo-v1-DiscountRule)
    - [Event](#buygo-v1-Event)
    - [EventItem](#buygo-v1-EventItem)
    - [GetEventRequest](#buygo-v1-GetEventRequest)
    - [GetEventResponse](#buygo-v1-GetEventResponse)
    - [GetMyRegistrationsRequest](#buygo-v1-GetMyRegistrationsRequest)
    - [GetMyRegistrationsResponse](#buygo-v1-GetMyRegistrationsResponse)
    - [ListEventRegistrationsRequest](#buygo-v1-ListEventRegistrationsRequest)
    - [ListEventRegistrationsResponse](#buygo-v1-ListEventRegistrationsResponse)
    - [ListEventsRequest](#buygo-v1-ListEventsRequest)
    - [ListEventsResponse](#buygo-v1-ListEventsResponse)
    - [ListManagerEventsRequest](#buygo-v1-ListManagerEventsRequest)
    - [ListManagerEventsResponse](#buygo-v1-ListManagerEventsResponse)
    - [RegisterEventRequest](#buygo-v1-RegisterEventRequest)
    - [RegisterEventResponse](#buygo-v1-RegisterEventResponse)
    - [RegisterItem](#buygo-v1-RegisterItem)
    - [Registration](#buygo-v1-Registration)
    - [UpdateEventRequest](#buygo-v1-UpdateEventRequest)
    - [UpdateEventResponse](#buygo-v1-UpdateEventResponse)
    - [UpdateEventStatusRequest](#buygo-v1-UpdateEventStatusRequest)
    - [UpdateEventStatusResponse](#buygo-v1-UpdateEventStatusResponse)
    - [UpdateRegistrationRequest](#buygo-v1-UpdateRegistrationRequest)
    - [UpdateRegistrationResponse](#buygo-v1-UpdateRegistrationResponse)
    - [UpdateRegistrationStatusRequest](#buygo-v1-UpdateRegistrationStatusRequest)
    - [UpdateRegistrationStatusResponse](#buygo-v1-UpdateRegistrationStatusResponse)
  
    - [EventStatus](#buygo-v1-EventStatus)
    - [RegistrationStatus](#buygo-v1-RegistrationStatus)
  
    - [EventService](#buygo-v1-EventService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="api_v1_auth-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/v1/auth.proto



<a name="buygo-v1-GetMeRequest"></a>

### GetMeRequest







<a name="buygo-v1-GetMeResponse"></a>

### GetMeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#buygo-v1-User) |  |  |






<a name="buygo-v1-ListAssignableManagersRequest"></a>

### ListAssignableManagersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| query | [string](#string) |  | Optional search filter by name or email. |






<a name="buygo-v1-ListAssignableManagersResponse"></a>

### ListAssignableManagersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| managers | [User](#buygo-v1-User) | repeated |  |






<a name="buygo-v1-ListUsersRequest"></a>

### ListUsersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  |  |
| page_token | [string](#string) |  |  |






<a name="buygo-v1-ListUsersResponse"></a>

### ListUsersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [User](#buygo-v1-User) | repeated |  |
| next_page_token | [string](#string) |  |  |






<a name="buygo-v1-LoginRequest"></a>

### LoginRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id_token | [string](#string) |  | Firebase ID token obtained from client-side authentication. |






<a name="buygo-v1-LoginResponse"></a>

### LoginResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| access_token | [string](#string) |  | Backend JWT for subsequent API calls. |
| user | [User](#buygo-v1-User) |  |  |






<a name="buygo-v1-UpdateUserRoleRequest"></a>

### UpdateUserRoleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| role | [UserRole](#buygo-v1-UserRole) |  |  |






<a name="buygo-v1-UpdateUserRoleResponse"></a>

### UpdateUserRoleResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#buygo-v1-User) |  |  |






<a name="buygo-v1-User"></a>

### User
User represents a platform user account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| email | [string](#string) |  |  |
| photo_url | [string](#string) |  |  |
| role | [UserRole](#buygo-v1-UserRole) |  |  |





 


<a name="buygo-v1-UserRole"></a>

### UserRole
User roles for role-based access control.

| Name | Number | Description |
| ---- | ------ | ----------- |
| USER_ROLE_UNSPECIFIED | 0 |  |
| USER_ROLE_USER | 1 | Regular user — can browse, place orders, and register for events. |
| USER_ROLE_CREATOR | 2 | Creator — can create and manage group buys and events. |
| USER_ROLE_SYS_ADMIN | 3 | System administrator — full platform access including user management. |


 

 


<a name="buygo-v1-AuthService"></a>

### AuthService
AuthService handles user authentication and user management.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Login | [LoginRequest](#buygo-v1-LoginRequest) | [LoginResponse](#buygo-v1-LoginResponse) | Login exchanges a Firebase ID token for a backend JWT. Access: Public |
| ListUsers | [ListUsersRequest](#buygo-v1-ListUsersRequest) | [ListUsersResponse](#buygo-v1-ListUsersResponse) | ListUsers returns a paginated list of all users. Access: Admin only |
| UpdateUserRole | [UpdateUserRoleRequest](#buygo-v1-UpdateUserRoleRequest) | [UpdateUserRoleResponse](#buygo-v1-UpdateUserRoleResponse) | UpdateUserRole changes a user&#39;s role. Access: Admin only |
| GetMe | [GetMeRequest](#buygo-v1-GetMeRequest) | [GetMeResponse](#buygo-v1-GetMeResponse) | GetMe returns the currently authenticated user&#39;s profile. Access: Authenticated |
| ListAssignableManagers | [ListAssignableManagersRequest](#buygo-v1-ListAssignableManagersRequest) | [ListAssignableManagersResponse](#buygo-v1-ListAssignableManagersResponse) | ListAssignableManagers returns users eligible to be assigned as managers (Creator or SysAdmin). Access: Creator |

 



<a name="api_v1_groupbuy-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/v1/groupbuy.proto



<a name="buygo-v1-AddProductRequest"></a>

### AddProductRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buy_id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| price_original | [int64](#int64) |  |  |
| exchange_rate | [double](#double) |  |  |
| rounding_config | [RoundingConfig](#buygo-v1-RoundingConfig) |  |  |
| specs | [string](#string) | repeated | Spec name strings to create as ProductSpec entries. |






<a name="buygo-v1-AddProductResponse"></a>

### AddProductResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| product | [Product](#buygo-v1-Product) |  |  |






<a name="buygo-v1-BatchUpdateStatusRequest"></a>

### BatchUpdateStatusRequest
BatchUpdateStatusRequest triggers FIFO batch status progression.
The system finds the oldest `count` items matching `spec_id` with status &lt; `target_status`,
and advances them to `target_status`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buy_id | [string](#string) |  |  |
| spec_id | [string](#string) |  | Filter by product spec. |
| target_status | [OrderItemStatus](#buygo-v1-OrderItemStatus) |  | Target status to advance items to. |
| count | [int32](#int32) |  | Number of items to process (oldest first). |






<a name="buygo-v1-BatchUpdateStatusResponse"></a>

### BatchUpdateStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| updated_count | [int32](#int32) |  | Actual number of items updated. |
| updated_order_ids | [string](#string) | repeated | Order IDs affected, for UI feedback. |






<a name="buygo-v1-CancelOrderRequest"></a>

### CancelOrderRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_id | [string](#string) |  |  |






<a name="buygo-v1-CancelOrderResponse"></a>

### CancelOrderResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_id | [string](#string) |  |  |
| status | [OrderItemStatus](#buygo-v1-OrderItemStatus) |  |  |






<a name="buygo-v1-Category"></a>

### Category
Category is a product category template with predefined spec names.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| spec_names | [string](#string) | repeated | Predefined spec names for products in this category. |






<a name="buygo-v1-ConfirmPaymentRequest"></a>

### ConfirmPaymentRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_id | [string](#string) |  |  |
| status | [PaymentStatus](#buygo-v1-PaymentStatus) |  | Must be CONFIRMED or REJECTED. |






<a name="buygo-v1-ConfirmPaymentResponse"></a>

### ConfirmPaymentResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_id | [string](#string) |  |  |
| status | [PaymentStatus](#buygo-v1-PaymentStatus) |  |  |






<a name="buygo-v1-CreateCategoryRequest"></a>

### CreateCategoryRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| spec_names | [string](#string) | repeated |  |






<a name="buygo-v1-CreateCategoryResponse"></a>

### CreateCategoryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| category | [Category](#buygo-v1-Category) |  |  |






<a name="buygo-v1-CreateGroupBuyRequest"></a>

### CreateGroupBuyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| products | [Product](#buygo-v1-Product) | repeated |  |
| cover_image_url | [string](#string) |  |  |
| deadline | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| shipping_configs | [ShippingConfig](#buygo-v1-ShippingConfig) | repeated |  |
| manager_ids | [string](#string) | repeated |  |
| exchange_rate | [double](#double) |  |  |
| rounding_config | [RoundingConfig](#buygo-v1-RoundingConfig) |  |  |
| source_currency | [string](#string) |  |  |






<a name="buygo-v1-CreateGroupBuyResponse"></a>

### CreateGroupBuyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buy | [GroupBuy](#buygo-v1-GroupBuy) |  |  |






<a name="buygo-v1-CreateOrderItem"></a>

### CreateOrderItem
CreateOrderItem specifies an item to include when creating or updating an order.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| product_id | [string](#string) |  |  |
| spec_id | [string](#string) |  |  |
| quantity | [int32](#int32) |  |  |
| status | [OrderItemStatus](#buygo-v1-OrderItemStatus) |  |  |






<a name="buygo-v1-CreateOrderRequest"></a>

### CreateOrderRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buy_id | [string](#string) |  |  |
| items | [CreateOrderItem](#buygo-v1-CreateOrderItem) | repeated |  |
| contact_info | [string](#string) |  |  |
| shipping_address | [string](#string) |  |  |
| shipping_method_id | [string](#string) |  |  |
| note | [string](#string) |  |  |






<a name="buygo-v1-CreateOrderResponse"></a>

### CreateOrderResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_id | [string](#string) |  |  |






<a name="buygo-v1-CreatePriceTemplateRequest"></a>

### CreatePriceTemplateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| source_currency | [string](#string) |  |  |
| exchange_rate | [double](#double) |  |  |
| rounding_config | [RoundingConfig](#buygo-v1-RoundingConfig) |  |  |






<a name="buygo-v1-CreatePriceTemplateResponse"></a>

### CreatePriceTemplateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template | [PriceTemplate](#buygo-v1-PriceTemplate) |  |  |






<a name="buygo-v1-DeletePriceTemplateRequest"></a>

### DeletePriceTemplateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template_id | [string](#string) |  |  |






<a name="buygo-v1-DeletePriceTemplateResponse"></a>

### DeletePriceTemplateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template_id | [string](#string) |  |  |






<a name="buygo-v1-GetGroupBuyRequest"></a>

### GetGroupBuyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buy_id | [string](#string) |  |  |






<a name="buygo-v1-GetGroupBuyResponse"></a>

### GetGroupBuyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buy | [GroupBuy](#buygo-v1-GroupBuy) |  |  |
| products | [Product](#buygo-v1-Product) | repeated |  |






<a name="buygo-v1-GetMyGroupBuyOrderRequest"></a>

### GetMyGroupBuyOrderRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buy_id | [string](#string) |  |  |






<a name="buygo-v1-GetMyGroupBuyOrderResponse"></a>

### GetMyGroupBuyOrderResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order | [Order](#buygo-v1-Order) |  | The user&#39;s order for this group buy, or null if none exists. |






<a name="buygo-v1-GetMyOrdersRequest"></a>

### GetMyOrdersRequest







<a name="buygo-v1-GetMyOrdersResponse"></a>

### GetMyOrdersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| orders | [Order](#buygo-v1-Order) | repeated |  |






<a name="buygo-v1-GetPriceTemplateRequest"></a>

### GetPriceTemplateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template_id | [string](#string) |  |  |






<a name="buygo-v1-GetPriceTemplateResponse"></a>

### GetPriceTemplateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template | [PriceTemplate](#buygo-v1-PriceTemplate) |  |  |






<a name="buygo-v1-GroupBuy"></a>

### GroupBuy
GroupBuy represents a group buying project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| cover_image_url | [string](#string) |  |  |
| status | [GroupBuyStatus](#buygo-v1-GroupBuyStatus) |  |  |
| exchange_rate | [double](#double) |  | Exchange rate from source currency to target currency. |
| rounding_config | [RoundingConfig](#buygo-v1-RoundingConfig) |  |  |
| source_currency | [string](#string) |  | Source currency code (e.g. &#34;JPY&#34;, &#34;USD&#34;). |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| deadline | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Order deadline — no new orders accepted after this time. |
| creator | [User](#buygo-v1-User) |  |  |
| managers | [User](#buygo-v1-User) | repeated |  |
| shipping_configs | [ShippingConfig](#buygo-v1-ShippingConfig) | repeated |  |






<a name="buygo-v1-ListCategoriesRequest"></a>

### ListCategoriesRequest







<a name="buygo-v1-ListCategoriesResponse"></a>

### ListCategoriesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| categories | [Category](#buygo-v1-Category) | repeated |  |






<a name="buygo-v1-ListGroupBuyOrdersRequest"></a>

### ListGroupBuyOrdersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buy_id | [string](#string) |  |  |






<a name="buygo-v1-ListGroupBuyOrdersResponse"></a>

### ListGroupBuyOrdersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| orders | [Order](#buygo-v1-Order) | repeated |  |






<a name="buygo-v1-ListGroupBuysRequest"></a>

### ListGroupBuysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  |  |
| page_token | [string](#string) |  |  |






<a name="buygo-v1-ListGroupBuysResponse"></a>

### ListGroupBuysResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buys | [GroupBuy](#buygo-v1-GroupBuy) | repeated |  |
| next_page_token | [string](#string) |  |  |






<a name="buygo-v1-ListManagerGroupBuysRequest"></a>

### ListManagerGroupBuysRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  |  |
| page_token | [string](#string) |  |  |






<a name="buygo-v1-ListManagerGroupBuysResponse"></a>

### ListManagerGroupBuysResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buys | [GroupBuy](#buygo-v1-GroupBuy) | repeated |  |
| next_page_token | [string](#string) |  |  |






<a name="buygo-v1-ListPriceTemplatesRequest"></a>

### ListPriceTemplatesRequest







<a name="buygo-v1-ListPriceTemplatesResponse"></a>

### ListPriceTemplatesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| templates | [PriceTemplate](#buygo-v1-PriceTemplate) | repeated |  |






<a name="buygo-v1-Order"></a>

### Order
Order represents a user&#39;s order within a group buy.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| group_buy_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| total_amount | [int64](#int64) |  | Total amount in target currency (smallest unit). |
| payment_status | [PaymentStatus](#buygo-v1-PaymentStatus) |  |  |
| payment_info | [PaymentInfo](#buygo-v1-PaymentInfo) |  |  |
| contact_info | [string](#string) |  | Sensitive fields — only visible to order owner and managers. |
| shipping_address | [string](#string) |  |  |
| note | [string](#string) |  |  |
| items | [OrderItem](#buygo-v1-OrderItem) | repeated |  |
| shipping_method_id | [string](#string) |  |  |
| shipping_fee | [int64](#int64) |  |  |






<a name="buygo-v1-OrderItem"></a>

### OrderItem
OrderItem represents a single line item in an order.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| product_id | [string](#string) |  |  |
| spec_id | [string](#string) |  |  |
| quantity | [int32](#int32) |  |  |
| status | [OrderItemStatus](#buygo-v1-OrderItemStatus) |  |  |
| product_name | [string](#string) |  | Snapshot fields — captured at order time for display consistency. |
| spec_name | [string](#string) |  |  |
| price | [int64](#int64) |  | Unit price at the time of order. |






<a name="buygo-v1-PaymentInfo"></a>

### PaymentInfo
PaymentInfo contains the payment proof submitted by the user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| method | [string](#string) |  | Payment method description (e.g. &#34;bank transfer&#34;, &#34;LINE Pay&#34;). |
| account_last5 | [string](#string) |  | Last 5 digits of the payment account for verification. |
| paid_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| amount | [int64](#int64) |  | Payment amount in target currency (smallest unit). |






<a name="buygo-v1-PriceTemplate"></a>

### PriceTemplate
PriceTemplate is a reusable pricing configuration for group buys.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| source_currency | [string](#string) |  |  |
| exchange_rate | [double](#double) |  |  |
| rounding_config | [RoundingConfig](#buygo-v1-RoundingConfig) |  |  |






<a name="buygo-v1-Product"></a>

### Product
Product represents a purchasable item within a group buy.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| group_buy_id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| description | [string](#string) |  |  |
| image_url | [string](#string) |  |  |
| price_original | [int64](#int64) |  | Original price in source currency (smallest unit). |
| exchange_rate | [double](#double) |  |  |
| rounding_config | [RoundingConfig](#buygo-v1-RoundingConfig) |  |  |
| price_final | [int64](#int64) |  | Final price after currency conversion and rounding (calculated server-side). |
| max_quantity | [int32](#int32) |  | Total available quantity (0 = unlimited). |
| specs | [ProductSpec](#buygo-v1-ProductSpec) | repeated | Product variants/options. |






<a name="buygo-v1-ProductSpec"></a>

### ProductSpec
ProductSpec represents a variant/option of a product (e.g. size, color).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |






<a name="buygo-v1-RoundingConfig"></a>

### RoundingConfig
RoundingConfig specifies how to round currency conversion results.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| method | [RoundingMethod](#buygo-v1-RoundingMethod) |  |  |
| digit | [int32](#int32) |  | Rounding digit: 0 = ones place, 1 = tens place, 2 = hundreds place. |






<a name="buygo-v1-ShippingConfig"></a>

### ShippingConfig
ShippingConfig defines a shipping option with its pricing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| type | [ShippingType](#buygo-v1-ShippingType) |  |  |
| price | [int64](#int64) |  | Shipping fee in the smallest currency unit (e.g. cents). |






<a name="buygo-v1-UpdateGroupBuyRequest"></a>

### UpdateGroupBuyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buy_id | [string](#string) |  |  |
| status | [GroupBuyStatus](#buygo-v1-GroupBuyStatus) |  |  |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| products | [Product](#buygo-v1-Product) | repeated |  |
| cover_image_url | [string](#string) |  |  |
| deadline | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| exchange_rate | [double](#double) |  |  |
| rounding_config | [RoundingConfig](#buygo-v1-RoundingConfig) |  |  |
| source_currency | [string](#string) |  |  |
| shipping_configs | [ShippingConfig](#buygo-v1-ShippingConfig) | repeated |  |
| manager_ids | [string](#string) | repeated |  |






<a name="buygo-v1-UpdateGroupBuyResponse"></a>

### UpdateGroupBuyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_buy | [GroupBuy](#buygo-v1-GroupBuy) |  |  |






<a name="buygo-v1-UpdateOrderRequest"></a>

### UpdateOrderRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_id | [string](#string) |  |  |
| items | [CreateOrderItem](#buygo-v1-CreateOrderItem) | repeated | Full replacement list of order items. |
| note | [string](#string) |  |  |






<a name="buygo-v1-UpdateOrderResponse"></a>

### UpdateOrderResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order | [Order](#buygo-v1-Order) |  |  |






<a name="buygo-v1-UpdatePaymentInfoRequest"></a>

### UpdatePaymentInfoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_id | [string](#string) |  |  |
| method | [string](#string) |  |  |
| account_last5 | [string](#string) |  |  |
| contact_info | [string](#string) |  |  |
| shipping_address | [string](#string) |  |  |
| paid_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| amount | [int64](#int64) |  |  |






<a name="buygo-v1-UpdatePaymentInfoResponse"></a>

### UpdatePaymentInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order | [Order](#buygo-v1-Order) |  |  |






<a name="buygo-v1-UpdatePriceTemplateRequest"></a>

### UpdatePriceTemplateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template_id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| source_currency | [string](#string) |  |  |
| exchange_rate | [double](#double) |  |  |
| rounding_config | [RoundingConfig](#buygo-v1-RoundingConfig) |  |  |






<a name="buygo-v1-UpdatePriceTemplateResponse"></a>

### UpdatePriceTemplateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template | [PriceTemplate](#buygo-v1-PriceTemplate) |  |  |





 


<a name="buygo-v1-GroupBuyStatus"></a>

### GroupBuyStatus
GroupBuyStatus represents the lifecycle of a group buy project.

| Name | Number | Description |
| ---- | ------ | ----------- |
| GROUP_BUY_STATUS_UNSPECIFIED | 0 |  |
| GROUP_BUY_STATUS_DRAFT | 1 |  |
| GROUP_BUY_STATUS_ACTIVE | 2 |  |
| GROUP_BUY_STATUS_ENDED | 3 |  |
| GROUP_BUY_STATUS_ARCHIVED | 4 |  |



<a name="buygo-v1-OrderItemStatus"></a>

### OrderItemStatus
OrderItemStatus tracks the fulfillment progress of an order item.
Progression: Unordered → Ordered → Arrived Overseas → Arrived Domestic → Ready for Pickup → Sent.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ITEM_STATUS_UNSPECIFIED | 0 |  |
| ITEM_STATUS_UNORDERED | 1 | User placed the item, not yet ordered from supplier. |
| ITEM_STATUS_ORDERED | 2 | Manager ordered from supplier. |
| ITEM_STATUS_ARRIVED_OVERSEAS | 3 | Item arrived at overseas warehouse. |
| ITEM_STATUS_ARRIVED_DOMESTIC | 4 | Item arrived at domestic warehouse. |
| ITEM_STATUS_READY_FOR_PICKUP | 5 | Item is ready for user pickup. |
| ITEM_STATUS_SENT | 6 | Item shipped to user or picked up. |
| ITEM_STATUS_FAILED | 7 | Item failed (out of stock, etc.). |



<a name="buygo-v1-PaymentStatus"></a>

### PaymentStatus
PaymentStatus tracks the payment verification state of an order.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PAYMENT_STATUS_UNSPECIFIED | 0 |  |
| PAYMENT_STATUS_UNSET | 1 | No payment info submitted yet. |
| PAYMENT_STATUS_SUBMITTED | 2 | User uploaded payment proof. |
| PAYMENT_STATUS_CONFIRMED | 3 | Manager confirmed the payment. |
| PAYMENT_STATUS_REJECTED | 4 | Manager rejected the payment. |



<a name="buygo-v1-RoundingMethod"></a>

### RoundingMethod
RoundingMethod defines how currency conversion results are rounded.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ROUNDING_METHOD_UNSPECIFIED | 0 |  |
| ROUNDING_METHOD_FLOOR | 1 |  |
| ROUNDING_METHOD_CEIL | 2 |  |
| ROUNDING_METHOD_ROUND | 3 |  |



<a name="buygo-v1-ShippingType"></a>

### ShippingType
ShippingType defines available shipping/pickup methods.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SHIPPING_TYPE_UNSPECIFIED | 0 |  |
| SHIPPING_TYPE_DELIVERY | 1 |  |
| SHIPPING_TYPE_STORE_PICKUP | 2 |  |
| SHIPPING_TYPE_MEETUP | 3 |  |


 

 


<a name="buygo-v1-GroupBuyService"></a>

### GroupBuyService
GroupBuyService manages group buying projects, products, orders, and pricing.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateGroupBuy | [CreateGroupBuyRequest](#buygo-v1-CreateGroupBuyRequest) | [CreateGroupBuyResponse](#buygo-v1-CreateGroupBuyResponse) | CreateGroupBuy creates a new group buy project. Access: Creator / Admin |
| ListGroupBuys | [ListGroupBuysRequest](#buygo-v1-ListGroupBuysRequest) | [ListGroupBuysResponse](#buygo-v1-ListGroupBuysResponse) | ListGroupBuys returns a paginated list of active group buys. Access: Public |
| ListManagerGroupBuys | [ListManagerGroupBuysRequest](#buygo-v1-ListManagerGroupBuysRequest) | [ListManagerGroupBuysResponse](#buygo-v1-ListManagerGroupBuysResponse) | ListManagerGroupBuys returns group buys managed by the current user. Access: Manager |
| GetGroupBuy | [GetGroupBuyRequest](#buygo-v1-GetGroupBuyRequest) | [GetGroupBuyResponse](#buygo-v1-GetGroupBuyResponse) | GetGroupBuy returns a group buy with its products. Access: Public |
| UpdateGroupBuy | [UpdateGroupBuyRequest](#buygo-v1-UpdateGroupBuyRequest) | [UpdateGroupBuyResponse](#buygo-v1-UpdateGroupBuyResponse) | UpdateGroupBuy updates group buy details, status, products, and managers. Access: Creator / Manager |
| AddProduct | [AddProductRequest](#buygo-v1-AddProductRequest) | [AddProductResponse](#buygo-v1-AddProductResponse) | AddProduct adds a product with specs to a group buy. Access: Creator / Manager |
| CreateCategory | [CreateCategoryRequest](#buygo-v1-CreateCategoryRequest) | [CreateCategoryResponse](#buygo-v1-CreateCategoryResponse) | CreateCategory creates a product category template. Access: Creator |
| ListCategories | [ListCategoriesRequest](#buygo-v1-ListCategoriesRequest) | [ListCategoriesResponse](#buygo-v1-ListCategoriesResponse) | ListCategories returns all product categories. Access: Public |
| CreateOrder | [CreateOrderRequest](#buygo-v1-CreateOrderRequest) | [CreateOrderResponse](#buygo-v1-CreateOrderResponse) | CreateOrder places a new order in a group buy. Access: Authenticated |
| CancelOrder | [CancelOrderRequest](#buygo-v1-CancelOrderRequest) | [CancelOrderResponse](#buygo-v1-CancelOrderResponse) | CancelOrder cancels an order (only if all items are still unordered). Access: Order Owner |
| GetMyGroupBuyOrder | [GetMyGroupBuyOrderRequest](#buygo-v1-GetMyGroupBuyOrderRequest) | [GetMyGroupBuyOrderResponse](#buygo-v1-GetMyGroupBuyOrderResponse) | GetMyGroupBuyOrder returns the current user&#39;s order for a specific group buy. Access: Authenticated |
| UpdateOrder | [UpdateOrderRequest](#buygo-v1-UpdateOrderRequest) | [UpdateOrderResponse](#buygo-v1-UpdateOrderResponse) | UpdateOrder modifies order items (only if all items are still unordered). Access: Order Owner |
| UpdatePaymentInfo | [UpdatePaymentInfoRequest](#buygo-v1-UpdatePaymentInfoRequest) | [UpdatePaymentInfoResponse](#buygo-v1-UpdatePaymentInfoResponse) | UpdatePaymentInfo submits or updates payment proof for an order. Access: Order Owner |
| GetMyOrders | [GetMyOrdersRequest](#buygo-v1-GetMyOrdersRequest) | [GetMyOrdersResponse](#buygo-v1-GetMyOrdersResponse) | GetMyOrders returns all orders placed by the current user. Access: Authenticated |
| BatchUpdateStatus | [BatchUpdateStatusRequest](#buygo-v1-BatchUpdateStatusRequest) | [BatchUpdateStatusResponse](#buygo-v1-BatchUpdateStatusResponse) | BatchUpdateStatus performs FIFO batch status progression on order items. Processes oldest items first, updating up to `count` items matching the spec. Access: Manager |
| ConfirmPayment | [ConfirmPaymentRequest](#buygo-v1-ConfirmPaymentRequest) | [ConfirmPaymentResponse](#buygo-v1-ConfirmPaymentResponse) | ConfirmPayment approves or rejects a user&#39;s payment. Access: Manager |
| ListGroupBuyOrders | [ListGroupBuyOrdersRequest](#buygo-v1-ListGroupBuyOrdersRequest) | [ListGroupBuyOrdersResponse](#buygo-v1-ListGroupBuyOrdersResponse) | ListGroupBuyOrders returns all orders for a group buy. Access: Manager |
| CreatePriceTemplate | [CreatePriceTemplateRequest](#buygo-v1-CreatePriceTemplateRequest) | [CreatePriceTemplateResponse](#buygo-v1-CreatePriceTemplateResponse) | CreatePriceTemplate creates a reusable pricing template. Access: Admin |
| ListPriceTemplates | [ListPriceTemplatesRequest](#buygo-v1-ListPriceTemplatesRequest) | [ListPriceTemplatesResponse](#buygo-v1-ListPriceTemplatesResponse) | ListPriceTemplates returns all pricing templates. Access: Admin |
| GetPriceTemplate | [GetPriceTemplateRequest](#buygo-v1-GetPriceTemplateRequest) | [GetPriceTemplateResponse](#buygo-v1-GetPriceTemplateResponse) | GetPriceTemplate returns a single pricing template. Access: Admin |
| UpdatePriceTemplate | [UpdatePriceTemplateRequest](#buygo-v1-UpdatePriceTemplateRequest) | [UpdatePriceTemplateResponse](#buygo-v1-UpdatePriceTemplateResponse) | UpdatePriceTemplate updates a pricing template. Access: Admin |
| DeletePriceTemplate | [DeletePriceTemplateRequest](#buygo-v1-DeletePriceTemplateRequest) | [DeletePriceTemplateResponse](#buygo-v1-DeletePriceTemplateResponse) | DeletePriceTemplate deletes a pricing template. Access: Admin |

 



<a name="api_v1_event-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/v1/event.proto



<a name="buygo-v1-CancelRegistrationRequest"></a>

### CancelRegistrationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registration_id | [string](#string) |  |  |






<a name="buygo-v1-CancelRegistrationResponse"></a>

### CancelRegistrationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registration_id | [string](#string) |  |  |
| status | [RegistrationStatus](#buygo-v1-RegistrationStatus) |  |  |






<a name="buygo-v1-CreateEventRequest"></a>

### CreateEventRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| start_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| end_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| discounts | [DiscountRule](#buygo-v1-DiscountRule) | repeated |  |
| items | [EventItem](#buygo-v1-EventItem) | repeated |  |
| location | [string](#string) |  |  |
| cover_image_url | [string](#string) |  |  |
| registration_deadline | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| payment_methods | [string](#string) | repeated |  |
| allow_modification | [bool](#bool) |  |  |
| manager_ids | [string](#string) | repeated |  |






<a name="buygo-v1-CreateEventResponse"></a>

### CreateEventResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#buygo-v1-Event) |  |  |






<a name="buygo-v1-DiscountRule"></a>

### DiscountRule
DiscountRule defines a quantity-based discount applied to registrations.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| min_quantity | [int32](#int32) |  | Minimum total items (sum across all selected items) to trigger the discount. |
| discount_amount | [int64](#int64) |  | Fixed discount amount off the total (e.g. 100 = 100 TWD off). |
| min_distinct_items | [int32](#int32) |  | Minimum number of distinct items required. |






<a name="buygo-v1-Event"></a>

### Event
Event represents a scheduled event with registration and pricing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| cover_image_url | [string](#string) |  |  |
| status | [EventStatus](#buygo-v1-EventStatus) |  |  |
| start_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| end_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| registration_deadline | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | No new registrations accepted after this time. |
| location | [string](#string) |  |  |
| creator | [User](#buygo-v1-User) |  |  |
| managers | [User](#buygo-v1-User) | repeated |  |
| payment_methods | [string](#string) | repeated | Accepted payment methods (e.g. &#34;bank transfer&#34;, &#34;LINE Pay&#34;). |
| items | [EventItem](#buygo-v1-EventItem) | repeated |  |
| allow_modification | [bool](#bool) |  | Whether registrants can modify their registration after submission. |
| discounts | [DiscountRule](#buygo-v1-DiscountRule) | repeated |  |






<a name="buygo-v1-EventItem"></a>

### EventItem
EventItem represents a selectable item within an event (e.g. a session, a ticket type).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| price | [int64](#int64) |  | Price per unit in the smallest currency unit. |
| min_participants | [int32](#int32) |  | Minimum participants required for this item to proceed. |
| max_participants | [int32](#int32) |  | Maximum participants allowed (0 = unlimited). |
| start_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Item availability time window. |
| end_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| allow_multiple | [bool](#bool) |  | Whether a user can select multiple quantities of this item. |






<a name="buygo-v1-GetEventRequest"></a>

### GetEventRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event_id | [string](#string) |  |  |






<a name="buygo-v1-GetEventResponse"></a>

### GetEventResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#buygo-v1-Event) |  |  |






<a name="buygo-v1-GetMyRegistrationsRequest"></a>

### GetMyRegistrationsRequest







<a name="buygo-v1-GetMyRegistrationsResponse"></a>

### GetMyRegistrationsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registrations | [Registration](#buygo-v1-Registration) | repeated |  |






<a name="buygo-v1-ListEventRegistrationsRequest"></a>

### ListEventRegistrationsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event_id | [string](#string) |  |  |






<a name="buygo-v1-ListEventRegistrationsResponse"></a>

### ListEventRegistrationsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registrations | [Registration](#buygo-v1-Registration) | repeated |  |






<a name="buygo-v1-ListEventsRequest"></a>

### ListEventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  |  |
| page_token | [string](#string) |  |  |






<a name="buygo-v1-ListEventsResponse"></a>

### ListEventsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| events | [Event](#buygo-v1-Event) | repeated |  |
| next_page_token | [string](#string) |  |  |






<a name="buygo-v1-ListManagerEventsRequest"></a>

### ListManagerEventsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  |  |
| page_token | [string](#string) |  |  |






<a name="buygo-v1-ListManagerEventsResponse"></a>

### ListManagerEventsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| events | [Event](#buygo-v1-Event) | repeated |  |
| next_page_token | [string](#string) |  |  |






<a name="buygo-v1-RegisterEventRequest"></a>

### RegisterEventRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event_id | [string](#string) |  |  |
| items | [RegisterItem](#buygo-v1-RegisterItem) | repeated |  |
| contact_info | [string](#string) |  | Contact information (e.g. phone, LINE ID). |
| notes | [string](#string) |  |  |






<a name="buygo-v1-RegisterEventResponse"></a>

### RegisterEventResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registration_id | [string](#string) |  |  |
| status | [RegistrationStatus](#buygo-v1-RegistrationStatus) |  |  |






<a name="buygo-v1-RegisterItem"></a>

### RegisterItem
RegisterItem specifies an event item and quantity to register for.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event_item_id | [string](#string) |  |  |
| quantity | [int32](#int32) |  | Number of units (usually 1, but can be more for guest tickets). |






<a name="buygo-v1-Registration"></a>

### Registration
Registration represents a user&#39;s registration for an event.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| event_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |
| status | [RegistrationStatus](#buygo-v1-RegistrationStatus) |  |  |
| payment_status | [PaymentStatus](#buygo-v1-PaymentStatus) |  |  |
| contact_info | [string](#string) |  |  |
| notes | [string](#string) |  |  |
| selected_items | [RegisterItem](#buygo-v1-RegisterItem) | repeated |  |
| user | [User](#buygo-v1-User) |  |  |
| total_amount | [int64](#int64) |  | Snapshot fields — captured at registration time. Total amount before discount. |
| discount_applied | [int64](#int64) |  | Discount amount applied based on DiscountRules. |






<a name="buygo-v1-UpdateEventRequest"></a>

### UpdateEventRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event_id | [string](#string) |  |  |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| start_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| end_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| location | [string](#string) |  |  |
| cover_image_url | [string](#string) |  |  |
| allow_modification | [bool](#bool) |  |  |
| items | [EventItem](#buygo-v1-EventItem) | repeated | Full replacement list of event items. |
| manager_ids | [string](#string) | repeated |  |
| discounts | [DiscountRule](#buygo-v1-DiscountRule) | repeated |  |






<a name="buygo-v1-UpdateEventResponse"></a>

### UpdateEventResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#buygo-v1-Event) |  |  |






<a name="buygo-v1-UpdateEventStatusRequest"></a>

### UpdateEventStatusRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event_id | [string](#string) |  |  |
| status | [EventStatus](#buygo-v1-EventStatus) |  |  |






<a name="buygo-v1-UpdateEventStatusResponse"></a>

### UpdateEventStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#buygo-v1-Event) |  |  |






<a name="buygo-v1-UpdateRegistrationRequest"></a>

### UpdateRegistrationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registration_id | [string](#string) |  |  |
| items | [RegisterItem](#buygo-v1-RegisterItem) | repeated |  |
| contact_info | [string](#string) |  |  |
| notes | [string](#string) |  |  |






<a name="buygo-v1-UpdateRegistrationResponse"></a>

### UpdateRegistrationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registration_id | [string](#string) |  |  |
| status | [RegistrationStatus](#buygo-v1-RegistrationStatus) |  |  |






<a name="buygo-v1-UpdateRegistrationStatusRequest"></a>

### UpdateRegistrationStatusRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registration_id | [string](#string) |  |  |
| status | [RegistrationStatus](#buygo-v1-RegistrationStatus) |  | Must be CONFIRMED or CANCELLED. |
| payment_status | [PaymentStatus](#buygo-v1-PaymentStatus) |  |  |






<a name="buygo-v1-UpdateRegistrationStatusResponse"></a>

### UpdateRegistrationStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registration_id | [string](#string) |  |  |
| status | [RegistrationStatus](#buygo-v1-RegistrationStatus) |  |  |
| payment_status | [PaymentStatus](#buygo-v1-PaymentStatus) |  |  |





 


<a name="buygo-v1-EventStatus"></a>

### EventStatus
EventStatus represents the lifecycle of an event.

| Name | Number | Description |
| ---- | ------ | ----------- |
| EVENT_STATUS_UNSPECIFIED | 0 |  |
| EVENT_STATUS_DRAFT | 1 |  |
| EVENT_STATUS_ACTIVE | 2 |  |
| EVENT_STATUS_ENDED | 3 |  |
| EVENT_STATUS_ARCHIVED | 4 |  |



<a name="buygo-v1-RegistrationStatus"></a>

### RegistrationStatus
RegistrationStatus tracks a user&#39;s registration state for an event.

| Name | Number | Description |
| ---- | ------ | ----------- |
| REGISTRATION_STATUS_UNSPECIFIED | 0 |  |
| REGISTRATION_STATUS_PENDING | 1 |  |
| REGISTRATION_STATUS_CONFIRMED | 2 |  |
| REGISTRATION_STATUS_CANCELLED | 3 |  |


 

 


<a name="buygo-v1-EventService"></a>

### EventService
EventService manages events, event items, and user registrations.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateEvent | [CreateEventRequest](#buygo-v1-CreateEventRequest) | [CreateEventResponse](#buygo-v1-CreateEventResponse) | CreateEvent creates a new event with items and discount rules. Access: Creator / Admin |
| ListEvents | [ListEventsRequest](#buygo-v1-ListEventsRequest) | [ListEventsResponse](#buygo-v1-ListEventsResponse) | ListEvents returns a paginated list of events. Access: Public |
| ListManagerEvents | [ListManagerEventsRequest](#buygo-v1-ListManagerEventsRequest) | [ListManagerEventsResponse](#buygo-v1-ListManagerEventsResponse) | ListManagerEvents returns events managed by the current user. Access: Manager |
| GetEvent | [GetEventRequest](#buygo-v1-GetEventRequest) | [GetEventResponse](#buygo-v1-GetEventResponse) | GetEvent returns a single event with its items and discount rules. Access: Public |
| RegisterEvent | [RegisterEventRequest](#buygo-v1-RegisterEventRequest) | [RegisterEventResponse](#buygo-v1-RegisterEventResponse) | RegisterEvent registers the current user for an event. Access: Authenticated |
| UpdateRegistration | [UpdateRegistrationRequest](#buygo-v1-UpdateRegistrationRequest) | [UpdateRegistrationResponse](#buygo-v1-UpdateRegistrationResponse) | UpdateRegistration modifies a user&#39;s registration items. Access: Registrant |
| UpdateRegistrationStatus | [UpdateRegistrationStatusRequest](#buygo-v1-UpdateRegistrationStatusRequest) | [UpdateRegistrationStatusResponse](#buygo-v1-UpdateRegistrationStatusResponse) | UpdateRegistrationStatus approves or rejects a registration. Access: Manager |
| CancelRegistration | [CancelRegistrationRequest](#buygo-v1-CancelRegistrationRequest) | [CancelRegistrationResponse](#buygo-v1-CancelRegistrationResponse) | CancelRegistration cancels a user&#39;s registration. Access: Registrant |
| GetMyRegistrations | [GetMyRegistrationsRequest](#buygo-v1-GetMyRegistrationsRequest) | [GetMyRegistrationsResponse](#buygo-v1-GetMyRegistrationsResponse) | GetMyRegistrations returns all registrations for the current user. Access: Authenticated |
| ListEventRegistrations | [ListEventRegistrationsRequest](#buygo-v1-ListEventRegistrationsRequest) | [ListEventRegistrationsResponse](#buygo-v1-ListEventRegistrationsResponse) | ListEventRegistrations returns all registrations for an event. Access: Manager |
| UpdateEvent | [UpdateEventRequest](#buygo-v1-UpdateEventRequest) | [UpdateEventResponse](#buygo-v1-UpdateEventResponse) | UpdateEvent updates event details, items, and managers. Access: Creator / Manager |
| UpdateEventStatus | [UpdateEventStatusRequest](#buygo-v1-UpdateEventStatusRequest) | [UpdateEventStatusResponse](#buygo-v1-UpdateEventStatusResponse) | UpdateEventStatus changes the event status (e.g. Draft → Active → Ended). Access: Creator / Manager |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

