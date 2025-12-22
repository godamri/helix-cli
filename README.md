Helix CLI
=========

> **Opinionated, Enterprise-Grade Microservice Generator for Go.**

Helix is not just a boilerplate generator. It is a strictly structured engineering standard enforcement tool. It scaffolds production-ready microservices implementing **Hexagonal Architecture**, **Transactional Outbox Pattern**, and **Observability** out of the box.

Stop wasting time debating folder structures. Focus on the Domain.

Why Helix?
----------

Most generators give you a "Hello World" HTTP server. **Helix gives you a survivable system.**

-   **Hexagonal Architecture:** Strict separation between `core` (domain), `adapter` (infra), and `port` (interfaces).

-   **Transactional Outbox:** Dual-write consistency (DB + Kafka) solved natively. No more distributed data inconsistencies.

-   **Dual Driver Strategy:** Choose between **Ent ORM** (Development Speed) or **Pgx/Raw SQL** (Performance/Control) per service.

-   **Ops-Ready:** Pre-configured with Docker Compose, Air (Hot Reload), Prometheus Metrics, OTel Tracing, and Structured Logging (slog).

-   **Database Migrations:** Integrated with [Atlas](https://atlasgo.io "null") for declarative schema management.

-   **gRPC & REST:** Dual transport layers generated automatically.

Installation
------------

```
go install github.com/godamri/helix-cli@latest

```

Ensure `$GOPATH/bin` is in your system `$PATH`.

Quick Start
-----------

### 1\. Initialize a New Service

Helix scaffolds a complete environment, including a `Makefile` that handles the entire lifecycle.

```
helix-cli init svc-payment

```

*You will be prompted to choose between `ent` (ORM) or `pgx` (Raw SQL).*

### 2\. Boot Up Infrastructure

Helix relies on Docker for dependencies (Postgres, Redpanda/Kafka, Redis).

```
cd svc-payment
make init

```

*This command will:*

1.  Spin up infra via Docker Compose.

2.  Wait for DB readiness.

3.  Apply initial migrations.

4.  Generate Ent/Proto code.

5.  Start the application with Hot Reload.

Usage
-----

### Scaffolding Entities

Don't write boilerplate by hand. Generate the Entity, Repository, Service, DTOs, and Handlers in one go.

```
helix-cli new entity Transaction

```

*This generates:*

-   `internal/core/entity/transaction.go`

-   `internal/core/port/transaction_repository.go`

-   `internal/adapter/repository/transaction_repository.go`

-   `internal/adapter/handler/v1/transaction_handler.go`

-   `ent/schema/transaction.go` (if using Ent)

> **Note:** Helix forces you to wire dependencies manually in `cmd/server/main.go`. **Explicit beats Implicit.**

### Adding Kafka Consumers

Generate a worker handler that fits the Hexagonal structure.

```
helix-cli new consumer UserCreated user.events.created

```

### Adding Redis Cache Repositories

Wrap your existing repositories with a caching layer.

```
helix-cli new cache Product

```

### Managing Migrations

Helix wraps `atlas` inside Docker to ensure consistency across team members.

```
# Create a new migration file based on Ent schema changes
helix-cli migrate diff add_user_columns

```

Architecture Overview
---------------------

Helix enforces a strict directory structure:

```
svc-payment/
├── cmd/
│   └── server/          # Entry point (Dependency Injection)
├── api/
│   └── proto/           # Protobuf definitions
├── ent/                 # Ent ORM generated code (if enabled)
├── internal/
│   ├── core/            # PURE DOMAIN LOGIC (No external deps)
│   │   ├── entity/      # Domain structs
│   │   ├── port/        # Interfaces (Service & Repository definitions)
│   │   ├── service/     # Business logic implementation
│   │   └── dto/         # Data Transfer Objects
│   ├── adapter/         # INFRASTRUCTURE LAYERS
│   │   ├── repository/  # DB implementations (Ent/SQL)
│   │   ├── handler/     # HTTP/gRPC handlers
│   │   └── worker/      # Kafka consumers & Outbox workers
│   └── pkg/             # Shared utilities (Config, Middleware)
├── migrations/          # Atlas SQL migrations
└── docker-compose.yml   # App & Infra definition

```

Configuration
-------------

Helix uses `envconfig` to load environment variables. Defaults are set in `internal/pkg/config/config.go`.

| Variable | Default | Description |
| --- |  --- |  --- |
| `APP_ENV` | `local` | Environment (local, dev, prod) |
| `DB_DSN` | \- | Postgres Connection String |
| `WORKER_CONCURRENCY` | `10` | Outbox worker parallelism |
| `AUTH_ENABLED` | `false` | Enable JWT/JWKS middleware |
| `AUDIT_OUTPUT` | `console` | Audit log destination (console/kafka) |

Contributing
------------

We value **pragmatism over purity**. If you have an improvement that makes the system more survivable or easier to operate, open a PR.

See [CONTRIBUTING.md](https://github.com/godamri/helix-cli/blob/main/CONTRIBUTING.md "null") for details.

License
-------

MIT © [Godamri](https://github.com/godamri/helix-cli "null")