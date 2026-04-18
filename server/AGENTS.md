# AI Agent Development Guide

## 1. Project Overview

- **Programming Language:** Golang
- **Architecture:** Clean Architecture
- **API:** gRPC with a gRPC-Gateway for RESTful JSON access.

## 2. Development Workflow

**This workflow is critically important and must be followed for all new implementations.**

0. **Understand the Ticket**
   * Check the contents of the given ticket and refer to `GEMINI.md` to understand the architecture and conventions.
   * Start implementation with a joyful spirit.
1. **Create a Development Branch**
   * Run `git fetch` to get the latest changes and create a new development branch from `origin/main`.
   * The branch name must comply with the **commitlint conventions**.
2. **Define the Domain**
   * Add the necessary domain definitions to `internal/domain/`.
   * Refer to existing definitions.
3. **Define Protobuf**
   * Add the necessary definitions to `proto/`.
   * Refer to existing `.proto` files.
4. **Define SQL**
   * Create SQL queries for `sqlc` in `internal/infrastructure/persistence/postgres/query/`.
5. **Generate Code**
   * Run `make gen` to generate code for `sqlc`, `pb`, `openapi`, etc.
6. **Implement the Repository**
   * Implement the repository in `internal/infrastructure/*` using the generated `sqlc` code.
7. **Implement the Usecase**
   * Implement the business logic in `internal/usecase/`.
8. **Implement the gRPC Handler**
   * Implement the gRPC handler in `internal/interface/grpc/`.
9. **Dependency Injection (DI)**
   * Add the dependencies of the implemented components to `di/di.go`.
10. **Confirm with the User**
    * Summarize the implementation details and present them to the user.
    * Confirm if it is okay to proceed with the subsequent work.
11. **Static Analysis**
    * Run `golangci-lint run ./...` to check the code quality.
12. **Testing**
    * Create tests corresponding to the implementation and run `go test ./...`.
    * Confirm that all tests pass.
13. **Commit & Push**
    * Run `git add`, `git commit`, and `git push`.
14. **Create a Pull Request**
    * Create a Pull Request on GitHub.
15. **Report Development Completion**
    * Report the completion of development to the user on Slack.

## 3. Clean Architecture

The project follows the principles of Clean Architecture to separate concerns and create a maintainable and testable codebase. The dependencies flow inwards:

`Interface` -> `Usecase` -> `Domain` <- `Infrastructure`

-   **`internal/domain`**: Contains the core business logic, entities, and interfaces for repositories. This layer is the heart of the application and has no external dependencies.
-   **`internal/usecase`**: Orchestrates the flow of data between the domain and the interfaces. It uses the domain services and repositories to perform application-specific tasks.
-   **`internal/interface`**: The outermost layer, responsible for handling external interactions. This includes gRPC handlers (`internal/interface/grpc`) and command-line interfaces (`internal/interface/cmd`). It calls the use cases to perform actions.
-   **`internal/infrastructure`**: Implements the interfaces defined in the domain layer, such as repositories for database access (`persistence`) or clients for external services (`oidc`). It handles all the details of data storage, external APIs, etc.

### Subdomain Granularity

Subdomains are created to encapsulate a cohesive set of functionalities and data within a larger domain. They promote better organization, reduce coupling, and improve maintainability.

-   **When to create a subdomain**: Consider creating a subdomain when a part of your domain has a distinct set of responsibilities, its own set of entities, and a clear boundary that separates it from the rest of the domain. For example, `form` and `submission` are closely related but `submission` has its own lifecycle and specific operations, making it a good candidate for a subdomain.
-   **Naming**: The main entity within a subdomain should be named after the subdomain itself (e.g., `submission.Submission`). Supporting entities within that subdomain should not repeat the subdomain name (e.g., `submission.History`, `submission.Value`).
-   **Dependencies**: Subdomains should primarily depend on their parent domain and other shared packages. Dependencies between sibling subdomains should be carefully managed to avoid circular dependencies.

## 4. Dependency Injection (DI)

Dependency Injection is managed by `go.uber.org/dig`.

-   **Configuration:** All dependencies are defined in `di/di.go`. This file wires together all the components of the application, from infrastructure implementations to use cases and interface handlers.
-   **Usage:** The DI container is built in `cmd/server/main.go` and `cmd/cli/main.go`, and then used to `Invoke` the top-level functions (e.g., the gRPC server or a CLI command handler).

## 5. Database Management

-   **Migrations:** Database schema migrations are handled by `github.com/golang-migrate/migrate`.
    -   Migration files (`.up.sql` and `.down.sql`) are located in `db/migrations/postgres/`.
    -   There is a `Makefile` in `db/migrations/` that may contain helper commands for creating new migration files.
-   **Seeding:** Seed data for development and testing is defined in `db/seeds/seed.yaml`. The logic for seeding is in `db/seeds/seeds.go`. The `dev-seed` make command can be used to run this.
-   **Queries**: SQL queries are written in `.sql` files within `internal/infrastructure/persistence/postgres/query/` and `sqlc` is used to generate type-safe Go code from them.

## 6. Code Generation

The project relies heavily on code generation to reduce boilerplate and ensure consistency.

-   **`make gen`**: This is the primary command to run all code generation steps.
-   **Protobuf & gRPC**:
    -   Protocol Buffer definitions (`.proto`) are located in the `proto/` directory.
    -   `buf` is used for linting, breaking change detection, and code generation. The configuration is in `buf.yaml` and `buf.gen.yaml`.
    -   `buf generate` creates Go gRPC servers/clients (`pb/`) and gRPC-Gateway stubs.
-   **OpenAPI**: `make gen` also generates OpenAPI v2 (Swagger) specifications from the gRPC definitions.
-   **SQLC**: As mentioned above, `sqlc` generates Go code from SQL queries.

## 7. Configuration

-   **Source:** Configuration is loaded from environment variables.
-   **File:** The structure of the configuration is defined in `config/config.go` using the `github.com/sethvargo/go-envconfig` library. This file shows all available configuration options (e.g., database connection, Keycloak settings, gRPC port).

## 8. Commands & Entrypoints

The `cmd/` directory contains multiple entrypoints for the application:

-   **`cmd/server/`**: The main entry point for running the gRPC and HTTP API server.
-   **`cmd/cli/`**: An entry point for command-line operations, such as running database seeds.
-   **`cmd/openapi223/`**: A utility to patch the generated OpenAPI spec.
-   **`cmd/proto2yaml/`**: A utility to convert protobuf service definitions into a YAML format.

## 9. Shared Packages (`pkg/`)

The `pkg/` directory contains utility packages that are shared across the project, such as:

-   `pkg/logger`: A structured logger.
-   `pkg/jwt`: JWT handling utilities.
-   `pkg/hash`: Hashing utilities.

## 10. Communication Flow

The server exposes both a gRPC and a RESTful JSON API on the same port using `cmux`.

1.  An incoming request hits the server.
2.  `cmux` determines if it's a gRPC request or an HTTP request.
3.  **gRPC requests** are routed directly to the gRPC server.
4.  **HTTP requests** are routed to the gRPC-Gateway, which translates the RESTful JSON request into a gRPC request and sends it to the gRPC server internally.

## 11. Authentication & Authorization (Auth)

Auth is handled via gRPC interceptors defined in `internal/interface/grpc/interceptor/interceptor.go`.

### Authentication (Who are you?)

-   **Provider:** Authentication is delegated to **Keycloak**.
-   **Flow:**
    1.  A client obtains a JWT access token from Keycloak.
    2.  The client sends a request to the API with the token in the `Authorization` header.
    3.  The `UnaryAuthInterceptor` or `StreamAuthInterceptor` intercepts the request.
    4.  It calls `token.Operator` (implemented using `gocloak`) to validate the token against Keycloak's public keys.
    5.  If the token is valid, it extracts the user's information (ID, email, etc.).
    6.  It calls `user.Service`'s `CreateIfNotExists` method to provision the user in the local database if this is their first time accessing the service.
    7.  The user's ID is injected into the request `context` for use in subsequent layers.

### Authorization (What can you do?)

-   **Implementation:** Authorization is a custom **Role-Based Access Control (RBAC)** implementation.
-   **Flow:**
    1.  The `UnaryAuthzInterceptor` or `StreamAuthzInterceptor` runs after the authentication interceptor.
    2.  It extracts the user's ID from the `context`.
    3.  It determines the required permission by parsing the gRPC method name (e.g., `/user.v1.UserService/GetMe` requires permission to perform the `GetMe` action on the `api.user.v1.UserService` resource).
    4.  It calls the `auth.Service.Enforce` method.
    5.  The `Enforce` method uses the `auth.Repository` to run a complex SQL query (`SelectUserAuthorizedPolices` in `internal/infrastructure/persistence/mysql/query/auth.sql`). This query joins the `users`, `user_groups`, `groups`, `group_roles`, `role_permissions`, `permissions`, and `resources` tables to determine all permissions the user has through their group memberships and assigned roles.
    6.  The service then checks if any of the user's permissions match the required permission for the action. Wildcards (`*`) are supported.
    7.  If a matching permission is found, the request is allowed to proceed; otherwise, a `Permission Denied` error is returned.

## 12. Development Commands (Makefile)

The `Makefile` defines various commands to streamline the development process.

- **`make init`**:
  - Uses `go install` to install the necessary tools for code generation and database migration:
    - `sqlc`: Generates Go code from SQL.
    - `migrate`: Database migration tool.
    - `protoc-gen-grpc-gateway`: Generates gRPC-Gateway code.
    - `protoc-gen-openapiv2`: Generates OpenAPI v2 definitions.
    - `protoc-gen-go`: Generates Go code from Protobuf.
    - `protoc-gen-go-grpc`: Generates gRPC server/client code.

- **`make dev`**:
  - Executes `docker compose up` to start the services defined in Docker Compose (application, database, etc.) locally.

- **`make dev-seed`**:
  - After generating code with `make gen`, it executes the `cli` command via `docker compose exec` to populate the database with seed data.

- **`make migrate`**:
  - Executes database migrations using the `migrate` CLI tool. The target database is specified within the Makefile.

- **`make migrate-create`** (inside `./db/migrations/Makefile`):
  - Interactively generates new migration files (`.sql`).

- **`make gen`**:
  - Comprehensively runs the project's code generation steps:
    - `sqlc generate`
    - `buf generate`
    - `go run cmd/openapi223/main.go`
    - `go run ./cmd/proto2yaml/proto_to_yaml.go`

- **`make pre-commit-install`**:
  - Sets up `pre-commit` and installs the Git commit hooks. This automatically runs linters and formatters before a commit.

## 13. Golang Implementation Guide

This section outlines the coding conventions and implementation patterns observed within the `internal` directory. Adhering to these guidelines will help maintain consistency and readability across the codebase.

### Naming Conventions

-   **Packages**: Package names are short, concise, and all lowercase (e.g., `user`, `auth`, `grpc`).
-   **Interfaces**:
    -   Domain-level interfaces for services or repositories follow the `type Name interface` pattern (e.g., `auth.Repository`, `auth.Service`).
    -   The main repository interface in a package should be named `Repository` (e.g., `user.Repository`, `submission.Repository`).
    -   Sub-repository interfaces should be named `[SubModel]Repository` (e.g., `submission.HistoryRepository`).
    -   The `sqlc` generated interface is named `Querier`.
-   **Structs**:
    -   Structs are named using `CamelCase` (e.g., `user.User`, `userHandler`).
    -   The main domain model struct in a package should be named the same as the package (e.g., `user.User`, `submission.Submission`).
    -   Sub-models within a package should not repeat the package name (e.g., `submission.History` instead of `submission.SubmissionHistory`).
    -   Infrastructure repositories are structs that implement a domain interface (e.g., `auth.Repository` struct).
    -   Usecase structs are suffixed with `UseCase` and are lowercase (e.g., `userUseCase`).
    -   gRPC handlers are suffixed with `Handler` and are lowercase (e.g., `userHandler`).
-   **Methods & Functions**:
    -   Public functions and methods use `CamelCase` (e.g., `NewUserUseCase`, `GetMe`).
    -   Private helper functions use `camelCase` (e.g., `userDomainToPb`).
-   **Variables**:
    -   Standard `camelCase` is used for local variables (e.g., `userId`, `pbItems`).
    -   Struct members are `CamelCase` (e.g., `user.User.Username`).

### Implementation Policies

-   **Constructor Functions**:
    -   Components are instantiated using `New...` functions (e.g., `NewService`, `NewRepository`, `NewUserUseCase`).
    -   These constructors receive dependencies as arguments and return an interface type, hiding the concrete implementation. This is central to the dependency injection pattern.
-   **Error Handling**:
    -   Errors from lower layers (like repositories) are generally not wrapped with additional context at the point of calling.
    -   In domain services, errors from repositories may be wrapped with `fmt.Errorf("..." %w", err)` to add business-contextual information.
    -   In gRPC handlers, errors are returned directly to be handled by interceptors or the gRPC framework, which will convert them to appropriate gRPC status codes.
-   **Context Propagation**:
    -   `context.Context` is the first parameter for all methods in the usecase, repository, and service layers.
    -   It is used for request-scoped values (like the user ID, injected in the auth interceptor via `contextx.WithUserID`), cancellation, and deadlines.
-   **Repository Pattern**:
    -   The `domain` layer defines the repository `interface` (e.g., `auth.Repository`).
    -   The `infrastructure` layer provides the concrete implementation (e.g., `auth.Repository` struct).
    -   Implementations in `infrastructure` typically hold a reference to the `sqlc.Queries` object.
-   **Usecase Layer**:
    -   The usecase struct holds references to its dependencies, which are domain repositories and services.
    -   It contains the core application logic, orchestrating calls to repositories to fulfill a specific task.
    -   For database transactions that involve multiple repository calls, the usecase can use a `trans.Repository` to execute operations within a transaction.
-   **gRPC Handlers**:
    -   The handler's responsibility is to:
        1.  Decode the gRPC request (`pb.Request`).
        2.  Call the appropriate usecase method.
        3.  Translate the domain model returned by the usecase into a gRPC response message (`pb.Response`).
    -   They contain minimal logic, focusing only on the translation between the transport layer (gRPC) and the application layer (usecase).
-   **CQRS for Complex Queries**:
    -   For complex read operations that involve multiple tables, aggregations, or require a specific data structure (Read Model) that differs from the domain model, it is acceptable to use a CQRS (Command Query Responsibility Segregation) approach.
    -   **Query-side**:
        -   Create a dedicated `Finder` or `QueryService` interface in the `usecase` layer (e.g., `user.Finder`).
        -   The implementation of this interface should be in the `infrastructure/persistence` layer.
        -   This service can directly return a Read Model (DTO) tailored for the specific view, bypassing the domain model and repository for read operations. This avoids bloating the domain model with properties that are only needed for display purposes.
        -   The query logic can be implemented using `sqlc` or a query builder like `goqu`.
    -   **Command-side**:
        -   The standard `Usecase` and `Repository` pattern should still be used for all command (create, update, delete) operations. This ensures that all business rules and invariants are enforced through the domain model.
    -   This approach provides a clear separation between read and write operations, improving performance and maintainability for complex queries while preserving the integrity of the domain model for command operations.

## 14. Git Commit Message Convention

- All commit messages must be in English.

## 15. Branching Strategy

All development should be done in a feature branch. Please follow the rules below when creating a branch.

1.  **Update Local Repository:** Before creating a new branch, update your local repository with the latest changes from the remote.
    ```shell
    git fetch origin
    ```
2.  **Create a New Branch:** Create a new branch from the `origin/main` branch.
    ```shell
    git checkout -b <branch-name> origin/main
    ```
3.  **Branch Naming Convention:** Branch names must follow the Conventional Commits specification, similar to commit messages. The format should be `type/short-description`.

    -   **`type`**: Must be one of the following: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`.
    -   **`short-description`**: A brief, hyphenated description of the branch's purpose in English.

    **Examples:**
    -   `feat/add-password-update-api`
    -   `fix/user-login-issue`
    -   `docs/update-readme`

## 16. MCP Server Integration

All interactions with GitHub, such as fetching issue details and creating pull requests, should be performed through the MCP server. Use the appropriate tools provided by the MCP server for these tasks.
