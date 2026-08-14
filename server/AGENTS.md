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
   * Give every new rpc a `hub.annotations.v1.method_rule` option with a
     `summary`. It is what the authorization interceptor, the `hub` CLI and the
     generated agent reference all read (see section 11).
   * Follow the API conventions in section 17.
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
-   **Web client**: `cmd/gen-web-client` writes `ui/web/src/api/operations.ts`, the verb/path table the browser client calls through, so the UI never repeats a REST path.
-   **Agent reference**: `hub api docs` writes `.agents/skills/hub-api/references/api-reference.md`, the API reference the `hub-api` agent skill ships.

    All three read `pkg/apicatalog`, so a proto change reaches the server, the web client, the CLI and the agent documentation in one `make gen`.

## 7. Configuration

-   **Source:** Configuration is loaded from environment variables.
-   **File:** The structure of the configuration is defined in `config/config.go` using the `github.com/sethvargo/go-envconfig` library. This file shows all available configuration options (e.g., database connection, Keycloak settings, gRPC port).
-   **Validation:** `EnvConfig.Validate` runs each section's own `Validate`. A new
    section's check has to be added there by hand - a promoted call would be
    ambiguous once more than one section has one, and a check nobody calls is a
    check that does not exist.

### AI chat backend

`chat.Service` has two implementations and `ANTHROPIC_MOCK` picks between them:

| Variable | Default | Meaning |
| --- | --- | --- |
| `ANTHROPIC_MOCK` | `false` | `true` swaps the Anthropic client for `internal/infrastructure/ai/mock`, which streams a scripted reply. `server/.env.local` sets it, so `make dev` needs no key and no network. |
| `ANTHROPIC_MOCK_DELAY` | `25ms` | Pause between streamed deltas, so the UI is seen filling in. |
| `ANTHROPIC_API_KEY` | - | Required unless `ANTHROPIC_MOCK=true`; the config refuses to start otherwise. |
| `ANTHROPIC_MODEL` | `claude-sonnet-4-6` | Model id, echoed by the mock so a screenshot says which backend answered. |
| `ANTHROPIC_SESSION_TOKEN_BUDGET` | `400000` | What one chat session may spend before it stops answering. Per session, so hitting it costs one conversation rather than the assistant. Zero or less removes the cap. |

Sending a message containing `!error` makes the mock fail the stream, which is
how the client's error path is exercised without breaking the real API, `!tool`
makes it call a tool, and `!write` makes it propose a change and wait for
approval (both below).

### Chat tools

`internal/infrastructure/ai/tool` lets the assistant answer from hub's own data
instead of from memory, and change it with the user's approval.

-   **What is exposed** is the `exposed` map in `toolbox.go`, written out by
    gRPC method path, with a flag saying whether each one writes. It is
    deliberately not a rule such as "anything called `Get*`", so adding an rpc
    to the API never silently hands it to a model.
-   **What is excluded** is the `escalation` map, and nothing on it reaches the
    model whatever the user is permitted to do and whatever they approve:

    ```
    AddPermissionsToRole -> AddRolesToGroup -> AddGroupsToUser
    ```

    Those three compose into a path to any permission at all, and every
    `Delete*` is excluded on the same grounds. Approval is not a sufficient
    guard for them: the threat is not that the model exceeds the user's rights -
    it cannot - but that text somebody else wrote, read back through a tool
    result, talks the model into proposing a change, and the person approving it
    is the person that text is written to persuade. Changing who can do what is
    rare, deliberate work with the highest cost when it goes wrong and the least
    to gain from automation, so it stays in the console. `New` refuses to build
    a tool box if an rpc appears in both maps.
-   **What a tool looks like** comes from `pkg/apicatalog`: the summary from the
    proto annotation becomes the description, and the request fields become the
    JSON Schema, constraints included.
-   **Who may call it** is checked twice. `Tools` filters the list with
    `auth.Service.Enforce` so the model is not offered work the user cannot do,
    and `Call` dispatches through `UnaryAuthzInterceptor`, which is what
    actually holds - a permission can be revoked between the two.
-   **How it is dispatched**: in-process, through the generated
    `ServiceDesc.Methods[i].Handler`, with the server's own authorization and
    validation interceptors passed in. There is therefore **one** implementation
    of "may this user do this" and **one** of "is this request well formed", and
    a tool call goes through both. No HTTP request is made and no bearer token
    is held, which is why this is not routed back through the REST gateway.
    **Do not reimplement `Enforce` or protovalidate here.**
-   **What comes back** is protojson with the `hidden` fields removed at any
    depth - `email` above all. Answering "who is in the admin group" needs a
    name, not an address, and the address is the field most directly usable for
    contacting or impersonating someone. Retention of what *is* sent is a
    separate question, settled by the Anthropic data-retention configuration.
-   **Prompt injection**: tool results carry text other users wrote - group
    descriptions, resource metadata, user names. The system prompt frames them
    as data rather than instructions. Keep that framing when editing it.

`maxToolRounds` in the Claude client bounds how many times the model may call
tools before it has to answer. Two rounds is the normal case ("find the group,
then read its members"), so a single round is not enough.

### The approval flow

A read runs the moment the model asks. **A change never does.** The model gets
as far as proposing one; the stream stops, and the change happens only if the
user says so.

-   **Why**, again, is not permissions. Every tool call goes through the same
    authorization the user's own requests do, so the model cannot exceed them.
    It is that tool results carry text other people wrote, and that text is in a
    position to talk the model into proposing something nobody asked for.
    Putting a person between the proposal and the change breaks that path -
    which is why the person is shown the **operation and its real arguments**,
    not a summary. "Add three users to a group" is not something anyone can
    meaningfully agree to.
-   **Why it is two rpcs**: grpc-gateway's server-streaming is one way. The
    answer stream has to end for the user to reply to it, so `SendMessage` stops
    on a `tool_proposal` frame and `ConfirmToolCall` opens a new stream that
    carries on. The paused conversation therefore has to outlive the request,
    which is what `chat_tool_proposals` is for.
-   **The continuation** in that row is opaque to everything but the backend
    that wrote it. Keeping it opaque is what stops a provider's message format
    becoming part of the schema; the Claude client stores its own small record
    rather than the SDK's parameter types, which marshal for sending but do not
    round trip.
-   **One approval is one change.** The decision is claimed with an update that
    matches only a pending row, so a retry, a double click or a replayed request
    finds nothing to claim. Only the first change in a turn is ever put to the
    user, so approving what you were shown cannot approve what you were not.
-   **What is recorded**: the change itself, in the audit log, against the
    person - with `channel = ai_chat`, the session id, and the `approval_id` of
    the proposal they approved. A decline is recorded on the proposal row.
-   **In the mock**: `!write` proposes a change and waits, so the card, the two
    buttons and both outcomes can be built and tested without an API key. The
    mock's `!tool` path deliberately picks a *read*, since that path runs the
    tool it picks.

`internal/infrastructure/ai/claude/pause_test.go` is the check that a tool
result carrying instructions cannot become a change: it feeds the injected text
in and asserts the change is proposed rather than run.

## 7.1 Audit log

`audit_logs` records every attempt to change the authorization graph: who made
it, by which route, what they asked for and whether it was allowed. Without it
there is no way to answer "why is this person in the admin group", which is the
question asked first whenever something is wrong.

-   **Where it is recorded** is `UnaryAuditInterceptor`, one interceptor rather
    than a line in each use case. Every path into the API - REST gateway, gRPC,
    the `hub` CLI, the assistant's tools - dispatches through the same chain, so
    there is nowhere for a change to go unrecorded. A use case that forgot to
    call a recorder would be silent, and silence is what the log has to rule
    out.
-   **The actor is always a person.** The assistant is a *channel*, never an
    actor: a tool call it makes is recorded against the user it was answering,
    with `channel = ai_chat` and the chat session id. "The AI did it" answers
    nothing, so nothing may record it that way.
-   **What is recorded** is the `audited` map in `interceptor/audit.go`, written
    out by gRPC method path, and `notAudited` holds the mutating rpcs
    deliberately left out with the reason. `UnclassifiedMutations` reports any
    mutating rpc in neither; `TestEveryMutationIsClassified` fails on one and the
    server logs one at startup. **Adding a mutating rpc means adding it to one
    of the two maps.**
-   **Refused attempts are recorded too**, which is why the interceptor runs
    *before* authorization. Someone trying to grant themselves admin and being
    told no is exactly the event worth having.
-   **A change and its record commit together.** The interceptor opens the
    transaction the use cases below join, so a failed record rolls the change
    back rather than leaving it unexplained. A failed *attempt* is recorded
    outside that transaction, since it was rolled back.
-   **Secrets never reach it**: the `secret` list is stripped from the recorded
    arguments at any depth. Add to that list, not to a per-message exception.
-   **Reading it** is `ListAuditLog` - `hub api call GET /api/v1/audit-logs`, or
    from the assistant, which has the same read-only tool. Filters: actor,
    resource, action, target, channel and a time range.
-   **Not covered**: `cli seed`, `cli resource-import`, migrations and hand-run
    SQL do not pass through gRPC and so are not recorded. They are operator
    actions on the box rather than API calls, and the deployment's own access
    controls are what accounts for them.

Adding the service means its RBAC resource has to exist: run `cli
resource-import` on an existing install, or nobody can be granted
`ListAuditLog`.

## 8. Commands & Entrypoints

The `cmd/` directory contains multiple entrypoints for the application:

-   **`cmd/server/`**: The main entry point for running the gRPC and HTTP API server.
-   **`cmd/cli/`**: Operational commands that talk to the database directly, such as running migrations and seeds.
-   **`cmd/hub/`**: The `hub` API client. Unlike `cmd/cli` it never touches the database; it calls the REST gateway with a token, and its command tree is generated from `pkg/apicatalog`. Install it with `make cli`.
-   **`cmd/openapi223/`**: A utility to patch the generated OpenAPI spec.
-   **`cmd/gen-web-client/`**: Writes the web client's operation table from the API catalog.

## 9. Shared Packages (`pkg/`)

The `pkg/` directory contains utility packages that are shared across the project, such as:

-   `pkg/logger`: A structured logger.
-   `pkg/jwt`: JWT handling utilities.
-   `pkg/hash`: Hashing utilities.
-   `pkg/apicatalog`: The API surface derived from the protobuf descriptors - REST mapping, request fields and RBAC rule per rpc. The authorization interceptor, the CLI and the code generators all read it.
-   `pkg/hubcli`: The `hub` CLI, built on `pkg/apicatalog`.

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
-   **Rules live in the proto.** Each rpc carries a `hub.annotations.v1.method_rule` option:

    ```proto
    rpc GetMe(GetMeRequest) returns (GetMeResponse) {
      option (google.api.http) = {get: "/api/v1/me"};
      option (hub.annotations.v1.method_rule) = {
        public: true
        summary: "Get the profile and groups of the authenticated user."
      };
    }
    ```

    -   `public: true` skips the enforcement step. The caller must still be
        authenticated, so public means "any signed-in user", not "anonymous".
        Only `GetMe` is public: a user with no roles yet must still be able to
        load their own profile.
    -   `resource` and `action` override what is enforced. Left empty they
        default to `api.<proto package>.<Service>` and the rpc name, which is
        what `cli resource-import` registers.
    -   `summary` is the one-line description the CLI and the generated agent
        reference show.

    `pkg/apicatalog` reads these options off the descriptors at startup, so
    making an endpoint public is a one-line proto change. Do not add a special
    case to the interceptor.
-   **Flow:**
    1.  The `UnaryAuthzInterceptor` or `StreamAuthzInterceptor` runs after the authentication interceptor.
    2.  It extracts the user's ID from the `context` and rejects an inactive account.
    3.  It looks the rpc up in `pkg/apicatalog` to get its resource, action and public flag. A method outside the catalog - gRPC health checking, say - falls back to the same naming convention.
    4.  Unless the rpc is public, it calls `auth.Service.Enforce`.
    5.  The `Enforce` method uses the `auth.Repository` to run a complex SQL query (`SelectUserAuthorizedPolices` in `internal/infrastructure/persistence/postgres/query/auth.sql`). This query joins the `users`, `user_groups`, `groups`, `group_roles`, `role_permissions`, `permissions`, and `resources` tables to determine all permissions the user has through their group memberships and assigned roles.
    6.  The service then checks if any of the user's permissions match the required permission for the action. `*` is a wildcard and may appear anywhere in a pattern, any number of times: `api.*`, `*Service` and `api.system.*.v1.*Service` are all valid.
    7.  If a matching permission is found, the request is allowed to proceed; otherwise, a `Permission Denied` error is returned.
-   **Policy cache:** `AUTHZ_POLICY_CACHE_TTL` (default 5m) memoises a user's
    effective policies, so a burst of requests runs the permission join once
    instead of once per call.

    The TTL does **not** decide how stale a decision can be. `rbac_revisions`
    holds a counter that database triggers bump on every write to `users`,
    `user_groups`, `groups`, `group_roles`, `roles`, `role_permissions`,
    `permissions` and `resources`. The cache re-reads that counter at most once
    a second and drops everything it holds when the number moves, so a grant or
    revocation lands within a second on every replica however long the TTL is.
    The TTL only bounds how long an entry nobody invalidated is kept.

    The triggers live in the database rather than the use cases because the
    graph is also edited by `cli seed`, `resource-import`, migrations and by
    hand; any of those forgetting to invalidate would leave a revoked
    permission working. **A new table that an authorization decision reads
    needs a trigger too.**

    `internal/infrastructure/auth/cache_integration_test.go` checks the
    triggers against a real database. It skips unless `HUB_TEST_DSN` is set.
-   **Inspecting permissions:** `hub api describe <rpc>` reports the resource
    and action an rpc needs, which is the quickest way to explain a 403.

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
  - Every tool is pinned to an explicit version through a `*_VERSION` variable
    at the top of the root `Makefile`. Do not switch them back to `@latest`:
    `make gen`'s output depends on the generator version, so an unpinned tool
    makes a one-line proto change rewrite the whole of `pb/`.
  - `PROTOC_GEN_GO_VERSION` tracks `google.golang.org/protobuf` and
    `GRPC_GATEWAY_VERSION` tracks `github.com/grpc-ecosystem/grpc-gateway/v2`,
    both in `go.mod`. When you bump either module, bump the matching variable
    and re-run `make gen` in the same change.

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

## 17. API Conventions

The CLI's commands, the web client's operation table and the RBAC verbs are all
generated from these definitions, so an inconsistency in a .proto becomes an
inconsistency in four places at once. Keep to the following.

### Naming

-   **CRUD** is `Get` / `List` / `Create` / `Update` / `Delete` followed by the
    entity: `GetUser`, `ListGroup`, `DeleteRole`.
-   **Many-to-many links** get exactly two rpcs, and no singular variants:
    -   `Add<Children>To<Parent>` - e.g. `AddRolesToGroup`, `AddUsersToGroup`
    -   `Remove<Children>From<Parent>` - e.g. `RemoveRolesFromGroup`
-   Do not add a "replace the whole set" rpc. One used to exist and its only
    caller used it to emulate a removal, which is what `Remove...` is for.

### Paths

-   `/api/v1/<plural-entity>` and `/api/v1/<plural-entity>/{id}` for CRUD.
-   `/api/v1/<plural-parent>/{id}/<plural-children>/add` and `.../remove` for
    links, always `POST`.
-   **The path parameter naming the entity is always `{id}`**, never
    `{group_id}` or `{role_id}`. The request field is `id` to match.
-   Lower case only. No camelCase segments.

### Requests and responses

-   A link request is `{ string id = 1; repeated string <child>_ids = 2; }`.
-   A mutation returns the entity it changed, so a caller does not have to
    re-read it. `AddUsersToGroupResponse` carries the updated `Group`.
-   Number fields from 1 in new messages. Where a number is skipped for
    compatibility, say so in a comment.

### Validation belongs in the proto

Request constraints are declared with [protovalidate](https://github.com/bufbuild/protovalidate)
and enforced by `UnaryValidateInterceptor`, which runs after authorization and
returns `InvalidArgument`. Do not re-check the same thing in a handler or a use
case: two places to state a rule is two places for them to disagree.

```proto
message CreateUserRequest {
  string username = 1 [(buf.validate.field).string = {min_len: 1, max_len: 64}];
  string email    = 2 [(buf.validate.field).string.email = true];
  repeated string group_ids = 3 [(buf.validate.field).repeated.items.string.uuid = true];
}
```

-   **Every id is a uuid.** A path parameter without that rule turns a typo into
    a confusing "not found"; a test enforces this.
-   **A `limit` is bounded** by `pagination.MaxLimit`, so a caller asking for
    more is told rather than quietly given less.
-   **A link request needs `min_items: 1`**: adding nothing to a group is a
    mistake, not a no-op.
-   Constrain an `optional` field freely - the rule is skipped when the field is
    absent, which is what "leave the password unchanged" relies on.

`pkg/apicatalog` reads the rules back off the descriptors, so `hub api describe`
and the flag help state them. **A rule added to a .proto reaches the CLI, the
agent reference and the server together.**

### Renaming an rpc is a data migration

The RBAC verb of an API permission *is* the rpc name. Renaming an rpc orphans
every permission naming it - the row survives, nothing enforces against it, and
the roles holding it silently lose the ability. Ship a migration that carries
the grants across; `000009_rename_api_permission_verbs` is the worked example.
