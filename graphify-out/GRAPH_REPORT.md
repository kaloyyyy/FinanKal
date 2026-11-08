# Graph Report - FinanKal  (2026-11-08)

## Corpus Check
- Corpus is ~33,069 words - fits in a single context window. You may not need a graph.

## Summary
- 1212 nodes · 2672 edges · 80 communities (43 shown, 37 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 85 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Go Engine Wiring
- gRPC Server Stubs
- gRPC Client Interfaces
- Docker Compose Services
- New Transaction Form
- Ledger Repository and Service
- Ledger Entries RPC
- Credit Statement RPC
- User Accounts RPC
- Project Docs and Roadmap
- UI Runtime Dependencies
- Cache Invalidation Keys
- Spring User Transactions
- Credit Statement Repository
- Spring DTO Mappers
- Decimal Credit Totals
- UI Dev Dependencies
- Create Credit Card RPC
- Spring Account Ledger Controllers
- Spring Transaction Requests
- Health Check RPC
- Pay Statement RPC
- Spring Credit Card API
- Recent Transactions RPC
- Get Transaction RPC
- Record Card Transaction RPC
- Billing Cycle Calculation
- User Transactions RPC
- Account Balance RPC
- Create Credit Card DTOs
- User Controller Endpoints
- Database Schema
- Account Summary RPC
- Create Transaction RPC
- ESLint Config
- Account Ledger UI
- TypeScript Config
- Protobuf Generated Init
- Transaction Entry Message
- Credit Card Service
- Recent Transaction Message
- Scheduled Transactions
- Global Exception Handling
- gRPC Client Config
- App Layout and Navbar
- Credit Entry Repository
- UI Transactions API
- Spring Transaction Service
- SvelteKit Build Tooling
- UI npm Scripts
- Dashboard Summary Cards
- UI Accounts API
- UI Accounts Loaders
- Create Transaction Entry Message
- Transaction Detail Page
- Transactions List Page
- Account DTO Mapping
- Spring App Config
- UI Account Store
- Financial Overview Cards
- Recent Transactions Widget
- Spring Boot Entrypoint
- Dashboard Page Loader
- Docker gRPC Config
- Local gRPC Config
- Prettier Config
- UI Static Docs
- Account Type Badge
- UI Transaction Types
- Accounts List Page
- SvelteKit App Types
- API Spring Maven Module
- Root Maven Module
- Go Module

## God Nodes (most connected - your core abstractions)
1. `file_finance_proto_rawDescGZIP()` - 35 edges
2. `GetUserTotalResponse` - 28 edges
3. `LedgerRepository` - 26 edges
4. `GetUserTotalRequest` - 25 edges
5. `TransactionResponse` - 22 edges
6. `Server` - 21 edges
7. `GetCreditCardStatementResponse` - 21 edges
8. `UnimplementedFinanceEngineServer` - 20 edges
9. `CreateCreditCardRequest` - 19 edges
10. `RecordCreditCardTransactionRequest` - 19 edges

## Surprising Connections (you probably didn't know these)
- `REST to gRPC Request Data Flow` --references--> `Go Engine Service (engine-go)`  [EXTRACTED]
  api-spring/OPENAPI.md → README.md
- `OpenAPI-First Design Pattern` --references--> `openapi.yaml Specification`  [EXTRACTED]
  README.md → api-spring/OPENAPI.md
- `OpenAPI-First Design Pattern` --cites--> `FinanKal REST API - OpenAPI-First Approach`  [EXTRACTED]
  README.md → api-spring/OPENAPI.md
- `OpenAPI-First Design Pattern` --references--> `Swagger UI`  [EXTRACTED]
  README.md → api-spring/OPENAPI.md
- `AccountController` --references--> `AccountService`  [EXTRACTED]
  api-spring/src/main/java/com/finankal/api/controller/AccountController.java → api-spring/src/main/java/com/finankal/api/service/AccountService.java

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **OpenAPI-First Code Generation Workflow** — api_spring_openapi_openapi_yaml, api_spring_openapi_generator_plugin, api_spring_openapi_apiapi, api_spring_openapi_generated_models [EXTRACTED 1.00]
- **Modules Feeding the Cash Forecast** — readme_salary_scheduler, readme_people_debt_module, readme_shared_expenses_module, readme_forecast_engine [EXTRACTED 1.00]
- **Controllers Implementing Generated ApiApi** — api_spring_openapi_accountcontroller, api_spring_openapi_transactioncontroller, api_spring_openapi_ledgerentrycontroller, api_spring_openapi_apiapi [EXTRACTED 1.00]

## Communities (80 total, 37 thin omitted)

### Community 0 - "Go Engine Wiring"
Cohesion: 0.11
Nodes (18): loadEnvFromRepoRoot(), main(), NewServer(), Exists(), Set(), getEnv(), getEnvInt(), NewRedis() (+10 more)

### Community 1 - "gRPC Server Stubs"
Cohesion: 0.07
Nodes (19): _FinanceEngine_CreateCreditCard_Handler(), _FinanceEngine_CreateTransaction_Handler(), _FinanceEngine_GetAccountSummary_Handler(), _FinanceEngine_GetBalance_Handler(), _FinanceEngine_GetCreditCardStatement_Handler(), _FinanceEngine_GetLedgerEntries_Handler(), _FinanceEngine_GetRecentTransactions_Handler(), _FinanceEngine_GetTransaction_Handler() (+11 more)

### Community 2 - "gRPC Client Interfaces"
Cohesion: 0.07
Nodes (7): NewFinanceEngineClient(), GetUserTotalRequest, GetUserTotalResponse, FinanceEngineClient, financeEngineClient, FinanceEngineServer, Server

### Community 3 - "Docker Compose Services"
Cohesion: 0.05
Nodes (40): build, container_name, depends_on, environment, ports, restart, context, dockerfile (+32 more)

### Community 4 - "New Transaction Form"
Cohesion: 0.05
Nodes (30): {
        canSubmit,
        loading = false,
        onCancel
    }, Props, Account, {
        accounts,
        entries = $bindable()
    }, creditCount, debitCount, Entry, Props (+22 more)

### Community 5 - "Ledger Repository and Service"
Cohesion: 0.09
Nodes (7): Get(), IsCacheMiss(), LedgerRepository, Service, NewService(), AccountSummary, TransactionDetail

### Community 6 - "Ledger Entries RPC"
Cohesion: 0.07
Nodes (3): GetLedgerEntriesRequest, GetLedgerEntriesResponse, LedgerEntry

### Community 7 - "Credit Statement RPC"
Cohesion: 0.06
Nodes (3): AccountType, GetCreditCardStatementRequest, GetCreditCardStatementResponse

### Community 8 - "User Accounts RPC"
Cohesion: 0.07
Nodes (3): GetUserAccountsRequest, GetUserAccountsResponse, Account

### Community 9 - "Project Docs and Roadmap"
Cohesion: 0.12
Nodes (25): AccountController, ApiApi Generated Interface, build-helper-maven-plugin, Generated Model Classes (com.finankal.api.generated.model), openapi-generator-maven-plugin, LedgerEntryController, openapi.yaml Specification, FinanKal REST API - OpenAPI-First Approach (+17 more)

### Community 10 - "UI Runtime Dependencies"
Cohesion: 0.10
Nodes (27): dependencies, axios@^1.19.0, chart.js@^4.5.1, class-variance-authority@^0.7.1, clsx@^2.1.1, dayjs@^1.11.21, decimal.js@^10.6.0, lucide-svelte@^1.0.1 (+19 more)

### Community 11 - "Cache Invalidation Keys"
Cohesion: 0.23
Nodes (19): CreditCardPaymentRequest, Delete(), Invalidate(), InvalidateAccount(), InvalidateCreditCard(), InvalidateCreditCardPayment(), InvalidateTransaction(), InvalidateUser() (+11 more)

### Community 12 - "Spring User Transactions"
Cohesion: 0.21
Nodes (4): TransactionDto, AccountMapper, AccountService, UserService

### Community 13 - "Credit Statement Repository"
Cohesion: 0.18
Nodes (3): CreditCardStatement, StatementRepository, StatementStatus

### Community 15 - "Decimal Credit Totals"
Cohesion: 0.11
Nodes (8): CreditCardBalance, CreditCardPayment, AccountType, Account, Entry, EntryType, RecentTransaction, TransactionEntry

### Community 16 - "UI Dev Dependencies"
Cohesion: 0.09
Nodes (23): devDependencies, prettier@^3.8.3, prettier-plugin-svelte@^4.1.0, prettier-plugin-tailwindcss@^0.8.0, svelte-check@^4.6.0, @sveltejs/vite-plugin-svelte@^7.1.2, tailwindcss@^4.3.0, @tailwindcss/container-queries@^0.1.1 (+15 more)

### Community 18 - "Spring Account Ledger Controllers"
Cohesion: 0.20
Nodes (4): AccountController, LedgerEntryController, LedgerEntryDto, LedgerEntryService

### Community 19 - "Spring Transaction Requests"
Cohesion: 0.15
Nodes (4): CreateTransactionEntryDto, CreateTransactionRequestDto, PayCreditCardStatementRequestDto, PayCreditCardStatementResponseDto

### Community 20 - "Health Check RPC"
Cohesion: 0.11
Nodes (3): file_finance_proto_rawDescGZIP(), HealthRequest, HealthResponse

### Community 22 - "Spring Credit Card API"
Cohesion: 0.16
Nodes (6): CreditCardController, CreditCardStatementDto, RecordCreditCardTransactionRequestDto, RecordCreditCardTransactionResponseDto, CreditCardMapper, CreditCardService

### Community 26 - "Billing Cycle Calculation"
Cohesion: 0.23
Nodes (14): BillingCycle, CreditCard, CreditCardTransactionRequest, CalculateBillingCycle(), CalculateDueDate(), CalculateStartDate(), CalculateStatementDate(), IsWithinBillingCycle() (+6 more)

### Community 30 - "User Controller Endpoints"
Cohesion: 0.16
Nodes (3): UserController, RecentTransactionDto, TransactionPageDto

### Community 31 - "Database Schema"
Cohesion: 0.24
Nodes (16): accounts, credit_card_entries, credit_card_statements, credit_cards, entries, idx_accounts_user_id, idx_credit_card_entries_entry, idx_credit_card_entries_statement (+8 more)

### Community 34 - "ESLint Config"
Cohesion: 0.13
Nodes (13): gitignorePath, eslint@^10.4.1, eslint-config-prettier@^10.1.8, @eslint/js@^10.0.1, eslint-plugin-svelte@^3.19.0, globals@^17.6.0, typescript-eslint@^8.60.1, eslint (+5 more)

### Community 35 - "Account Ledger UI"
Cohesion: 0.21
Nodes (8): { balance }, Props, { account }, {entries}, Props, LedgerEntryDto, currency, { data }

### Community 36 - "TypeScript Config"
Cohesion: 0.14
Nodes (13): compilerOptions, allowJs, checkJs, esModuleInterop, forceConsistentCasingInFileNames, moduleResolution, resolveJsonModule, rewriteRelativeImportExtensions (+5 more)

### Community 39 - "Credit Card Service"
Cohesion: 0.26
Nodes (3): CardRepository, Service, NewService()

### Community 41 - "Scheduled Transactions"
Cohesion: 0.26
Nodes (3): ScheduledTransaction, ScheduledTransactionOccurrence, ScheduleRepository

### Community 44 - "App Layout and Navbar"
Cohesion: 0.20
Nodes (5): Favicon (Svelte logo SVG), Svelte Logo, mobileOpen, navItems, { children }

### Community 46 - "UI Transactions API"
Cohesion: 0.29
Nodes (7): getTransaction(), getUserTransactions(), Transaction, TransactionEntry, TransactionsResponse, load(), load()

### Community 48 - "SvelteKit Build Tooling"
Cohesion: 0.22
Nodes (8): @sveltejs/adapter-auto@^7.0.1, @sveltejs/kit@^2.63.0, @tailwindcss/vite@^4.3.0, vite@^8.0.16, @sveltejs/adapter-auto, @sveltejs/kit, @tailwindcss/vite, vite

### Community 49 - "UI npm Scripts"
Cohesion: 0.22
Nodes (9): scripts, build, check, check:watch, dev, format, lint, prepare (+1 more)

### Community 50 - "Dashboard Summary Cards"
Cohesion: 0.22
Nodes (5): Props, {
        title,
        value,
        description,
        icon,
        variant = "default"
    }, Props, {
        totalCredit = 0,
        totalDebit = 0,
        netWorth = 0
    }, { data }

### Community 51 - "UI Accounts API"
Cohesion: 0.44
Nodes (6): Account, getAccount(), getAccountBalance(), getLedgerEntries(), api(), load()

### Community 52 - "UI Accounts Loaders"
Cohesion: 0.31
Nodes (5): getUserAccounts(), createTransaction(), load(), actions, load()

### Community 54 - "Transaction Detail Page"
Cohesion: 0.25
Nodes (5): balanced, currency, { data }, totalCredit, totalDebit

### Community 57 - "Spring App Config"
Cohesion: 0.29
Nodes (6): name, port, grpc, client, spring, application

### Community 58 - "UI Account Store"
Cohesion: 0.33
Nodes (5): svelte@^5.56.1, account, loading, AccountDto, svelte

### Community 59 - "Financial Overview Cards"
Cohesion: 0.33
Nodes (4): Props, {
        totalCredit,
        totalDebit,
        netWorth
    }, Props, {
        title,
        value,
        type="positive"
    }

### Community 62 - "Dashboard Page Loader"
Cohesion: 0.70
Nodes (3): getRecentTransactions(), getUserFinancialSummary(), load()

### Community 63 - "Docker gRPC Config"
Cohesion: 0.50
Nodes (3): host, grpc, client

### Community 64 - "Local gRPC Config"
Cohesion: 0.50
Nodes (3): host, grpc, client

### Community 67 - "UI Static Docs"
Cohesion: 0.67
Nodes (3): finankal-ui SvelteKit Project, SvelteKit HTML Shell Template, Allow-All Crawl Policy

## Knowledge Gaps
- **13 isolated node(s):** `com.finankal:api-spring`, `github.com/kaloy/finankal/engine-go`, `./.svelte-kit/tsconfig.json`, `com.finankal:finankal`, `Sprint Changelog (SP0-SP3)` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 420 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **37 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `GetUserTotalResponse` connect `gRPC Client Interfaces` to `Account Summary RPC`, `Create Transaction RPC`, `gRPC Server Stubs`, `Protobuf Generated Init`, `Create Transaction Entry Message`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **What connects `com.finankal:api-spring`, `github.com/kaloy/finankal/engine-go`, `./.svelte-kit/tsconfig.json` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Go Engine Wiring` be split into smaller, more focused modules?**
  _Cohesion score 0.11210317460317461 - nodes in this community are weakly interconnected._
- **Why does `RecentTransaction` connect `Recent Transaction Message` to `Account Summary RPC`, `Create Transaction RPC`, `Create Transaction Entry Message`, `Protobuf Generated Init`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Should `gRPC Server Stubs` be split into smaller, more focused modules?**
  _Cohesion score 0.07215541165587419 - nodes in this community are weakly interconnected._
- **Why does `GetUserTotalRequest` connect `gRPC Client Interfaces` to `Account Summary RPC`, `Create Transaction RPC`, `gRPC Server Stubs`, `Protobuf Generated Init`, `Create Transaction Entry Message`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Should `gRPC Client Interfaces` be split into smaller, more focused modules?**
  _Cohesion score 0.06866002214839424 - nodes in this community are weakly interconnected._