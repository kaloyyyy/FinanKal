# FinanKal

FinanKal is a modular financial system project built with:

- **Go** – core engine and gRPC server  
- **Spring Boot** – REST API for clients  
- **PostgreSQL** – database  
- **Docker & Docker Compose** – for local development and containerization  
- **Protobuf / gRPC** – for inter-service communication

---

## Repository Structure

```
FinanKal/
├─ engine-go/         # Go engine service (gRPC)
├─ api-spring/        # Spring Boot REST API
├─ proto/             # Protobuf definitions
├─ infra/             # Docker Compose, infrastructure setup
└─ README.md          # Project documentation
```

---

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)  
- [Java 21](https://adoptium.net/)  
- [Maven 3.9+](https://maven.apache.org/download.cgi)  
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)  
- IntelliJ IDEA (optional, for IDE support)

---

## Local Setup

### 1. Database

Start PostgreSQL via Docker Compose:

```bash
cd infra
docker-compose up -d postgres
````

**Environment variables:**

* DB_USER: `finance`
* DB_PASSWORD: `finance123`
* DB_NAME: `finance_db`
* DB_PORT: `5432`

---

### 2. Go Engine Service

Build and run the Go gRPC server:

```bash
cd engine-go
go mod tidy
go build -o server ./cmd/server
./server
```

**gRPC default port:** `50051`

---

### 3. Spring Boot REST API

Build and run the Spring Boot API:

```bash
cd api-spring
mvn clean package -DskipTests
java -jar target/api-spring-0.0.1-SNAPSHOT.jar
```

**REST API default port:** `8080`
**Connects to:** Go engine gRPC service on `50051`

---

### 4. Running Everything with Docker

From `infra/`:

```bash
docker-compose up --build
```

Services included:

* `finance_postgres` – PostgreSQL
* `finance_go_engine` – Go gRPC engine
* `finance_spring_api` – Spring Boot API

---

## Protobuf / gRPC

All `.proto` files live in the `proto/` folder.

* Use `protoc` to generate Go and Java stubs.
* Ensure `engine-go/go.mod` and `api-spring/pom.xml` reference the generated files.

Example Go generation:

```bash
protoc --go_out=. --go-grpc_out=. proto/*.proto
```

---

## REST API - OpenAPI First Approach

The Spring Boot API follows an **OpenAPI-first** design pattern:

### Key Features
- **Single Source of Truth**: OpenAPI specification (`api-spring/src/main/resources/openapi.yaml`)
- **Auto-Generated Code**: Controllers, models, and interfaces are generated from the spec
- **Swagger UI**: Interactive API documentation at `http://localhost:8080/swagger-ui.html`
- **Type-Safe Models**: Generated DTOs with validation annotations

### Workflow
1. **Design**: Update `openapi.yaml` with new endpoints
2. **Generate**: Run `mvn clean compile` to generate code
3. **Implement**: Add business logic in controller classes
4. **Test**: Use Swagger UI or the provided `requests.http` file

For detailed OpenAPI documentation, see [api-spring/OPENAPI.md](api-spring/OPENAPI.md)

---

## Go Module Tips

* Ensure `GOPATH` and `GOROOT` are correctly configured for IntelliJ.
* Avoid replacing modules with paths inside `$GOPATH/pkg/mod` — use a local folder if needed:

```go
replace github.com/aws/aws-sdk-go-v2 => ../local-aws-sdk-go-v2
```

* Run:

```bash
go mod tidy
```

to synchronize dependencies.

---

## IntelliJ Setup

* Open monorepo as a single project.
* Enable Go Modules in `engine-go/go.mod`.
* For Spring Boot, open as a Maven project.
* Synchronize dependencies after changes in `go.mod` or `pom.xml`.

---

## Troubleshooting

* **Port conflicts**: check `5432` for PostgreSQL (`sudo lsof -i :5432`)
* **gRPC server not found**: ensure `server` binary exists in `engine-go`
* **IntelliJ “Cannot resolve module”**: run `go mod tidy` and synchronize in IDE

---

## Sprint changelog

This changelog summarizes sprint-scoped commits in the repository (labels: SP0, SP1, SP2, SP3). Entries list commit hash, date, and commit message.

### SP0 (2026-03-17)
- a809b2d4ea3368d1ed23cf6ed77f42e0ad2c86d3 — 2026-03-17 — SP0 : setup env.
- 66c5f7e5d5597314740be06996909a163afdfdc5 — 2026-03-15 — Initial commit - FinanKal clean structure

### SP1 (2026-04-01 — 2026-04-03)
- b057338cbcbeefdee4f106ad781a696b634eaef1 — 2026-04-01 — SP1 feat: implement gRPC endpoints in engine-go
- aaf7b56682f5eefb59e60187cd7ca34cd1dfdfa3 — 2026-04-01 — SP1 chore: rename springboot main application
- dc7c96dd23d43da6ee62982be7f3e0a91dc0fce8 — 2026-04-03 — SP1 update proto file
- 065554a1abdc5a5cc8f4944497b4cf113286c4ab — 2026-04-03 — SP1 implement springboot http endpoints
- 0ff68e101429631356ddf5ac02a553d7e8024f23 — 2026-04-03 — SP1 update engine-go to accommodate spring boot
- c872298ea35616fced5f2b9759e862a72f8da902 — 2026-04-03 — SP1 openapi first approach

### SP2 (2026-04-06)
- 5ba28ae9a071b0d9ec4f77327a64e3d3fb500e52 — 2026-04-06 — SP2 (API): add get user balances
- 38a936fd2cfcff57221158670df9234f29ca5b11 — 2026-04-06 — SP2 (core): add user balances
- 59c3417fb3a0cdd3a5ea5b4b71cffb2fa417d01d — 2026-04-06 — SP2 (core): add cache to user balances
- f69b38171c316fdc89980745544590ecb13508b5 — 2026-04-06 — SP2 (core): add logging for cache hits
- bbfdc5bedd20232e788dc6b66546290b49ed25fe — 2026-04-06 — SP2 IDE settings

### SP3 (2026-05-06 — 2026-08-05)
- 1c3aea41537257208d827afe29480089ed457561 — 2026-05-06 — SP3 (core): logging
- cf322289c1081006192725379567830d3dd8a900 — 2026-05-06 — SP3 (core): update redis
- 45424c160488d9f4abd2e53d0eb39e96d01cfa99 — 2026-08-05 — SP3 (db): formatting

---

## FinanKal Sprint Plan (Updated)

Sprint 0 — Foundation & Project Setup

Goal: Have a running environment with all services.

Tasks:

- Monorepo skeleton (engine-go, api-spring, proto, infra)
- Docker Compose with PostgreSQL, Go engine, Spring Boot API, Redis
- Go gRPC HealthCheck service
- Spring Boot REST API HealthCheck
- gRPC proto setup

Deliverable:

- Run docker-compose up → Go engine + Spring API + PostgreSQL + Redis running
- Git repo initialized

Sprint 1 — Ledger Core

Goal: Build the financial foundation.

Go Engine Tasks:

- Accounts: Cash, CreditCard, Income, Expense, Receivable
- Transactions and double-entry ledger entries
- Compute balances per account
- Redis caching for balances
- gRPC endpoints for ledger operations

Spring Boot Tasks:

- REST endpoints for accounts, transactions, ledger entries
- Connect Spring Boot to Go engine via gRPC

Deliverable:

- Record transactions, compute balances
- REST + gRPC fully wired
- Redis cache integrated for fast balance queries

Sprint 2 — Credit Card Module

Goal: Track credit card purchases and cycles.

Go Engine Tasks:

- Credit card accounts
- Record credit card transactions (debit expenses, credit liability)
- Statement & due date logic

Spring Boot Tasks:

- REST endpoints for credit card CRUD
- Endpoint to record card transaction

Deliverable:

- Record card purchase
- Calculate statement date and due date
- Ledger entries auto-created
- Redis cache updated on card-related ledger changes

Sprint 3 — Salary Scheduler

Goal: Automatic income events.

Go Engine Tasks:

- Salary schedule table
- Auto-generate future salary events
- Update forecast cache when salary added

Spring Boot Tasks:

- REST API: Add/Update salary schedule

Deliverable:

- Salary events auto-inserted on scheduled dates
- Forecast reflects salary inflow

Sprint 4 — Cash Ledger & Forecast Engine

Goal: Centralize cash movements & daily cash forecast.

Go Engine Tasks:

- Compute daily forecast: current balance + future inflows/outflows
- Redis caching for forecast results
- Update cache when transactions, salary, or debts change

Spring Boot Tasks:

- Endpoint /forecast?days=90
- Endpoint /forecast/daily?range

Deliverable:

- Daily cash forecast available
- Cached in Redis for fast API responses

Sprint 5 — People / Debt Module

Goal: Track friends/people who owe you.

Go Engine Tasks:

- People table
- Receivable/Payable accounts for each person
- Ledger entries for loans and repayments
- Cache person balances in Redis

Spring Boot Tasks:

- REST APIs: /people, /loan, /settle, /people/{id}/balance

Deliverable:

- Track loans and repayments
- Redis cache keeps balances fast
- Forecast considers expected repayments

Sprint 6 — Shared Expenses Module

Goal: Split expenses among people using credit card.

Go Engine Tasks:

- Record shared expense
- Ledger entries: split between your portion and receivables
- Update forecast and cache

Spring Boot Tasks:

- REST API: /shared-expense

Deliverable:

- Shared expense recorded automatically
- Ledger reflects your cost + friends’ debt
- Forecast updated

Sprint 7 — Forecast Integration with Debts

Goal: Integrate friend debts and shared expenses into forecast.

Go Engine Tasks:

- Include expected inflows/outflows from debts in daily forecast
- Cache invalidation logic for forecast when debts change

Spring Boot Tasks:

- No new endpoints — existing /forecast now includes debts

Deliverable:

- Accurate cash forecast including debts and shared expenses

Sprint 8 — Testing & Quality (JaCoCo)

Goal: Make system stable and maintainable.

Tasks:

- Unit tests for Go engine
- Integration tests for REST API → gRPC → DB → Redis
- Test credit card cycles, salary, debts, shared expenses
- Measure code coverage (JaCoCo for Spring Boot)

Deliverable:

- ≥80% coverage
- All edge cases tested

Sprint 9 — Observability & Metrics

Goal: Monitor system health.

Tasks:

- Add Prometheus metrics: transactions, balances, gRPC latency
- Grafana dashboard
- Redis hit/miss metrics

Deliverable:

- System monitoring dashboards running

Sprint 10 — Deployment & Hosting

Goal: Make the system production-ready.

Tasks:

- Deploy on your Ryzen 9 PC or RPi
- Use Docker Compose in production mode
- Optional: Cloudflare tunnel for remote access
- Nginx reverse proxy

Deliverable:

- Fully running FinanKal system accessible remotely
- Database, Go engine, Spring API, Redis all running in production mode

Sprint 11 — Advanced Features / Optimizations

Goal: Prepare for future enhancements.

Tasks:

- Improve Redis cache logic (e.g., TTL for forecasts, batch invalidation)
- Optimize gRPC calls for high traffic
- Add historical reports (monthly summary, friend debt history)
- Add user authentication & JWT

Deliverable:

- Optimized system ready for extended features

Key Notes for Redis Integration
Cache balances, forecasts, person debts in Go engine.
Invalidate cache on ledger write (transaction, salary, debt, shared expense).
Spring Boot API should not cache business logic, only pass through Go engine responses.
