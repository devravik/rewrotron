# BEHAVIOR-MODEL.md — Three-Tier IR & Canonical Behavior Specification

> **Authority Level:** Supreme Behavioral Contract (Critical Foundation Document #1)  
> **Status:** Authoritative Core Specification  
> **Target Audience:** All AI Coding Agents, Target Adapters, and Verifiers  

---

## 1. The Three-Tier Intermediate Representation (IR)

To enable true framework-neutral reconstruction, Rewrotron models applications through three progressive tiers of abstraction:

```text
┌────────────────────────────────────────────────────────────────────────┐
│ 1. SOURCE IR (What the code is syntactically)                          │
│    • Concrete syntax trees, symbol declarations, class hierarchies     │
│    • Raw file layouts, legacy framework helper invocations             │
│    • Declared and implicit routing configuration                       │
├────────────────────────────────────────────────────────────────────────┤
│ 2. APPLICATION IR (What the application is structurally)               │
│    • Entrypoints, HTTP route handlers, background tasks, cron jobs     │
│    • Structural services, controllers, data models, entity schemas     │
│    • Shared state, dependency call graphs, transactional workflows     │
├────────────────────────────────────────────────────────────────────────┤
│ 3. BEHAVIOR IR (What the outside world observes)                       │
│    • Framework-independent operations and behavioral contracts         │
│    • Preconditions (database fixtures, authenticated session context)   │
│    • Input schemas (parameters, headers, payloads, file uploads)       │
│    • State transitions & database mutations (INSERT/UPDATE/DELETE)     │
│    • External side effects (emails, webhooks, message queues)          │
│    • Response specifications (status codes, redirects, body schemas)   │
│    • Dynamic variability normalization rules                           │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Canonical Behavior IR Hierarchy

```text
Application
    └── EntryPoint (HTTP Route, CLI Command, Cron Job, Event Consumer)
          └── Workflow (Optional multi-step stateful sequence)
                └── Operation (Single atomic request/response interaction)
                      ├── Preconditions (Database state, session state, environment)
                      ├── Inputs (Headers, query params, body params, uploaded files)
                      ├── Authentication & Authorization (Roles, permissions, session keys)
                      ├── Business Rules (Validation, calculations, conditional branches)
                      ├── Database Effects (Tables mutated, columns updated, row counts)
                      ├── External Side Effects (Emails, webhook dispatches, filesystem writes)
                      ├── Response (Status code, redirect rules, body schema, headers)
                      └── Variability Rules (Dynamic token normalizations)
```

---

## 3. Canonical Operation Specification

Every migrated operation or endpoint in Behavior IR is formally defined by a YAML/JSON structure following this exact canonical schema:

```yaml
operation_id: "order-checkout-001"
version: "1.0"
description: "Process cart checkout, deduct inventory, and record new customer order"

entrypoint:
  type: "http"
  method: "POST"
  path: "/checkout/process"
  content_types:
    - "application/x-www-form-urlencoded"
    - "application/json"

authentication:
  required: true
  mechanism: "session"
  session_keys:
    - "user_id"
    - "logged_in"
  unauthorized_response:
    status: 302
    redirect_path: "/login"

authorization:
  roles: ["customer", "admin"]
  permissions: ["orders.create"]
  forbidden_response:
    status: 403
    body_contains: "Unauthorized action"

preconditions:
  database:
    - table: "users"
      condition: "id = {session.user_id} AND active = 1"
    - table: "carts"
      condition: "user_id = {session.user_id} AND item_count > 0"
    - table: "inventory"
      condition: "product_id IN ({cart.product_ids}) AND quantity >= {cart.quantities}"

inputs:
  parameters:
    - name: "cart_id"
      type: "integer"
      required: true
      source: "body"
      validation: "exists:carts,id"
    - name: "payment_method"
      type: "string"
      required: true
      source: "body"
      allowed_values: ["credit_card", "paypal", "store_credit"]
    - name: "shipping_address_id"
      type: "integer"
      required: true
      source: "body"
  headers:
    - name: "X-Requested-With"
      required: false
      expected: "XMLHttpRequest"

business_rules:
  - id: "BR-001"
    description: "Calculate 8% state sales tax on taxable line items"
  - id: "BR-002"
    description: "Apply $10 discount if coupon 'WELCOME10' is active in session"
  - id: "BR-003"
    description: "Atomic transaction wrap: fail all inserts if inventory deduction fails"

database_effects:
  transactional: true
  mutations:
    - action: "INSERT"
      table: "orders"
      columns:
        user_id: "{session.user_id}"
        total_amount: "{computed.total}"
        status: "'pending'"
        created_at: "{dynamic.timestamp}"
      captured_keys:
        created_order_id: "LAST_INSERT_ID"
    - action: "INSERT"
      table: "order_items"
      batch_from: "{cart.items}"
      columns:
        order_id: "{captured.created_order_id}"
        product_id: "{item.product_id}"
        quantity: "{item.quantity}"
        unit_price: "{item.price}"
    - action: "UPDATE"
      table: "inventory"
      where: "product_id = {item.product_id}"
      mutation: "quantity = quantity - {item.quantity}"
    - action: "DELETE"
      table: "cart_items"
      where: "cart_id = {input.cart_id}"

session_effects:
  flash_messages:
    - key: "success"
      value: "Your order has been placed successfully."
  mutations:
    - action: "UNSET"
      key: "active_cart_id"

external_effects:
  emails:
    - recipient: "{user.email}"
      template: "order_confirmation"
      sandbox_policy: "mock_and_record"
  webhooks: []
  filesystem: []

response:
  status: 302
  headers:
    Location: "/orders/{captured.created_order_id}"
  redirect:
    exact_path: "/orders/{captured.created_order_id}"
    query_params: {}
  body:
    type: "empty"

variability:
  dynamic_fields:
    - path: "response.headers.Set-Cookie"
      type: "session_cookie"
      pattern: "[a-zA-Z0-9_-]+=[a-zA-Z0-9]+"
      normalization_rule: "mask"
    - path: "database_effects.orders.created_at"
      type: "iso8601_or_unix"
      normalization_rule: "timestamp_window_60s"
```

---

## 4. Multi-Target Portability of Behavior IR

Because Behavior IR describes **pure external observations and state transitions**, it serves as the single source of truth for all target adapters:
- **Laravel Target Adapter:** Maps inputs to Form Requests, database effects to Eloquent/Transactions, and responses to Redirects/Blade views.
- **Go Target Adapter:** Maps inputs to request DTO structs with validator tags, database effects to sqlc / database transactions, and responses to HTTP 302 handlers.
- **Rust Target Adapter:** Maps inputs to Serde-deserialized structs, database effects to SQLx transactions, and responses to Axum response tuples.
