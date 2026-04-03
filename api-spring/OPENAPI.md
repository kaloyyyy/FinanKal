# FinanKal REST API - OpenAPI-First Approach

## Overview

The FinanKal REST API uses an **OpenAPI-First** approach, where the API specification is the source of truth. The Java code (controllers, models, and interfaces) is automatically generated from the OpenAPI specification.

## Architecture

### OpenAPI Specification
- **Location**: `src/main/resources/openapi.yaml`
- **Format**: OpenAPI 3.0.0
- **Defines**: All API endpoints, request/response models, and documentation

### Generated Code
The Maven build automatically generates:
1. **API Interface** (`com.finankal.api.generated.ApiApi`)
   - Interface with all endpoint methods
   - Swagger annotations for documentation
   
2. **Model Classes** (`com.finankal.api.generated.model.*`)
   - DTOs for request/response payloads
   - Bean validation annotations
   - Jackson serialization configuration

### Implementation
Custom controllers implement the generated interfaces to provide business logic:
- `AccountController` 
- `TransactionController`
- `LedgerEntryController`

## API Endpoints

### Accounts
- `GET /api/accounts/{id}` - Get account summary
- `GET /api/accounts/{id}/balance` - Get account balance

### Transactions
- `POST /api/transactions` - Create a new transaction

### Ledger Entries
- `GET /api/ledger-entries/{accountId}` - Get ledger entries for an account

## Workflow: Modifying the API

### To Add a New Endpoint:

1. **Update OpenAPI Specification**
   ```yaml
   paths:
     /api/new-endpoint:
       get:
         operationId: newEndpoint
         tags: [NewFeature]
         # Define parameters, responses, etc.
   ```

2. **Regenerate Code**
   ```bash
   mvn clean compile
   ```

3. **Implement the Generated Interface**
   - Update or create the controller class
   - Implement the new method from the generated interface
   - Add business logic

### To Modify an Existing Endpoint:

1. **Update OpenAPI Specification** with parameter/response changes
2. **Run Maven**
   ```bash
   mvn clean compile
   ```
3. **Update Controller Implementation** to match new signature

### To Add New Model Fields:

1. **Update the Component Schema** in `openapi.yaml`
   ```yaml
   components:
     schemas:
       MyModel:
         properties:
           newField:
             type: string
   ```
2. **Regenerate**
   ```bash
   mvn clean compile
   ```

## Build Configuration

### Maven Plugins

**OpenAPI Generator Plugin** (`openapi-generator-maven-plugin`)
- Generates API interfaces and models from `openapi.yaml`
- Configuration:
  - Generator: `spring` (Spring Boot compatible)
  - Package: `com.finankal.api.generated`
  - Options: Bean validation, Spring Boot 3, interface-only mode

**Build Helper Plugin** (`build-helper-maven-plugin`)
- Adds generated sources to compile path
- Combines protobuf and OpenAPI generated sources

### Dependencies

Required dependencies for OpenAPI generated code:
- `jackson-databind-nullable` - Handles nullable fields
- Spring Boot Web & Validation
- OpenAPI annotations

## Development Workflow

### 1. Design the API
Edit `openapi.yaml` with new endpoints and schemas

### 2. Generate Code
```bash
mvn clean compile
```

### 3. Implement Controllers
```java
@RestController
public class MyController implements ApiApi {
    @Override
    public ResponseEntity<MyDto> myEndpoint(...) {
        // Implementation
    }
}
```

### 4. Test with Swagger UI
- Start the application
- Visit `http://localhost:8080/swagger-ui.html`
- Test endpoints directly from the UI

## Benefits of OpenAPI-First

✅ **Single Source of Truth** - Specification drives code generation  
✅ **Consistency** - API matches spec exactly  
✅ **Documentation** - Auto-generated from spec  
✅ **Type Safety** - Generated models with validation  
✅ **Swagger UI** - Interactive API documentation  
✅ **Contract Testing** - Validate requests/responses  
✅ **Reduced Boilerplate** - No manual endpoint routing  

## Swagger UI

Access the interactive API documentation:
```
http://localhost:8080/swagger-ui.html
```

Or the OpenAPI JSON endpoint:
```
http://localhost:8080/v3/api-docs
```

## Current Implementation

### Controllers Implementing Generated Interface

All controllers are designed to implement from the generated `ApiApi` interface:
- Each method signature matches the OpenAPI spec
- Swagger annotations are inherited from interface
- Logging is added for monitoring
- Business logic delegates to services

### Data Flow

```
REST Request
    ↓
@RestController (implements ApiApi)
    ↓
Service Layer (gRPC client)
    ↓
Go Engine (gRPC Server)
    ↓
Database/Cache
    ↓
Response DTO
```

## Troubleshooting

**Missing Generated Classes?**
```bash
mvn clean compile
```

**OpenAPI Spec Parse Error?**
- Validate YAML syntax: `yamllint openapi.yaml`
- Check schema definitions
- Ensure proper indentation

**Jackson Serialization Issues?**
- Ensure `jackson-databind-nullable` dependency is present
- Check model annotations

## References

- [OpenAPI 3.0 Specification](https://spec.openapis.org/oas/v3.0.0)
- [OpenAPI Generator](https://openapi-generator.tech/)
- [Spring OpenAPI Docs](https://springdoc.org/)

