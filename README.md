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

