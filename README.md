# Bitcoin Wallet

A self-custodial Bitcoin wallet built on **multi-party computation (MPC)** — no single party ever holds a complete private key. The backend is a set of microservices that cooperate to generate keys, derive addresses, index the chain, and sign and broadcast transactions. Currently running against a Bitcoin **regtest** node for development.

> ⚠️ **Status: early development.** Core backend services are functional, but the project is not feature-complete and **not intended for production use**. The frontend, deployment, and additional services are planned (see [Roadmap](#roadmap)).

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![.NET](https://img.shields.io/badge/.NET-8.0-512BD4?style=for-the-badge&logo=dotnet&logoColor=white)
![Java](https://img.shields.io/badge/Java-21-ED8B00?style=for-the-badge&logo=openjdk&logoColor=white)
![Angular](https://img.shields.io/badge/Angular-21-DD0031?style=for-the-badge&logo=angular&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

---

## Table of Contents

- [Architecture](#architecture)
- [Components](#components)
- [Repository Layout](#repository-layout)
- [Implemented Features](#implemented-features)
- [Getting Started](#getting-started)
- [Database Migrations](#database-migrations)
- [Environment Variables](#environment-variables)
- [Ports](#ports)
- [API Overview](#api-overview)
- [Frontend](#frontend)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Architecture

```
                                ┌─────────────────────┐
                                │     Frontend        │  Angular 21 (in progress)
                                └──────────┬──────────┘
                                           │ HTTPS
                                ┌──────────▼──────────┐
                                │     API Gateway     │  .NET 8 · Ocelot · JWT
                                │      (:8080)        │
                                └──────────┬──────────┘
                                           │ routes (Ocelot + Eureka service discovery)
              ┌────────────────────────────┼─────────────────────────────┐
              │                            │                             │
   ┌──────────▼──────────┐    ┌────────────▼──────────┐     ┌────────────▼────────────┐
   │   Wallet Service    │    │ Transaction Service   │     │ Blockchain Indexing     │
   │   .NET 8 (:8081)    │    │ .NET 8 (:8082)        │     │ Go (:8083 REST/:50052   │
   │   wallets, keys     │    │ transactions, signing │     │ gRPC)                   │
   └──────────┬──────────┘    └────────────┬──────────┘     │ UTXO store, chain state │
              │                            │                └────────────┬────────────┘
              │ gRPC                       │ gRPC                        │ JSON-RPC
   ┌──────────▼────────────────────────────▼─────────────────────────────▼──────────┐
   │                                                                                │
   │  ┌──────────────────────┐   ┌───────────────────────┐   ┌──────────────────┐   │
   │  │    MPC Service       │   │   Service Registry    │   │ bitcoind         │   │
   │  │    Go · tss-lib      │   │   Eureka (:8761)      │   │ (regtest)        │   │
   │  │    (:50051)          │   │                       │   │ (:18443)         │   │
   │  │    DKG + signing     │   └───────────────────────┘   └──────────────────┘   │
   │  └──────────┬───────────┘                                                      │
   │  │    MPC Service       │   │   Service Registry    │   │ bitcoind         │   │
   │  │    Go · tss-lib      │   │   Eureka (:8761)      │   │ (regtest)        │   │
   │  │    (:50051)          │   │                       │   │ (:18443)         │   │
   │  │    DKG + signing     │   └───────────────────────┘   └──────────────────┘   │
   │  └──────────┬───────────┘                                                      │
   │             │                                                                  │
   └─────────────┼──────────────────────────────────────────────────────────────────┘
                 │
   ┌─────────────▼──────────────────────────────────────────────────────────────┐
   │                         PostgreSQL 16 (:5433)                              │
   │            wallets · addresses · transactions · UTXOs · chain state        │
   └────────────────────────────────────────────────────────────────────────────┘

   Kafka + Schema Registry + AKHQ (docker-compose) — provisioned, will carry
   events between services once the User / Notification services are added.
```

**Key design points**

- **No single point of key custody** — private keys never exist in one place. Wallet keys are generated via a distributed key generation (DKG) protocol and transactions are signed with threshold signatures across MPC parties (`tss-lib`).
- **Service discovery** — all services register with **Eureka** (Spring Cloud Netflix) and are located by name through the gateway.
- **gRPC for inter-service calls** — Wallet ↔ MPC (key generation), Transaction ↔ MPC (signing), Transaction ↔ Indexer (chain state).
- **REST for client-facing APIs** — routed through the API Gateway at `/api/v1/*`.
- **Event-driven (planned)** — services will communicate over Kafka topics (wallet events, transaction confirmations, notifications); the `UserService` and `NotificationService` are next on the roadmap.

---

## Components

| Component | Stack | Port(s) | Status |
|---|---|---|---|
| `ServiceRegistry` | Java 21 · Spring Boot · Netflix Eureka Server | `8761` | ✅ Implemented |
| `ApiGateway` | .NET 8 · Ocelot · Steeltoe Eureka · JWT Bearer | `8080` | ✅ Implemented |
| `WalletService` | .NET 8 · EF Core · MediatR · gRPC | `8081` (REST), `50053` (gRPC) | ✅ Implemented |
| `TransactionService` | .NET 8 · EF Core · NBitcoin · gRPC | `8082` | ✅ Implemented |
| `BlockchainIndexingService` | Go · btcd · chi · pgx · gRPC | `8083` (REST), `50052` (gRPC) | ✅ Implemented |
| `MPC` | Go · bnb-chain/tss-lib v2 · gRPC | `50051` | ✅ Implemented |
| `Frontend` | Angular 21 · TypeScript | `4200` | 🚧 In progress (scaffold) |
| `PostgreSQL` | Postgres 16 (docker) | `5433` (host) | ✅ Provisioned |
| `Kafka / Schema Registry / AKHQ` | Confluent 7.6 (docker) | `9092` / `3000` / `3001` | 🚧 Provisioned, integration planned |
| `UserService` | .NET 8 · EF Core · Kafka | - | 🚧 Planned |
| `NotificationService` | .NET 8 · Kafka | - | 🚧 Planned |

<!-- TODO: add rows here as new services are implemented; update statuses as work lands -->

---

## Repository Layout

```
bitcoin-wallet/
├── backend/
│   ├── NetSolution/                    # .NET solution: gateway + REST services
│   │   ├── ApiGateway/                 # Ocelot gateway, JWT auth, Eureka
│   │   ├── WalletService/              # wallet creation, address derivation
│   │   └── TransactionService/         # transactions, fee estimation, signing
│   ├── ServiceRegistry/                # Spring Boot Eureka server
│   ├── MPC/                            # Go MPC service: DKG + threshold signing
│   │   ├── proto/                      # gRPC contracts (wallet, transaction)
│   │   ├── cmd/server/                 # gRPC server entrypoint
│   │   └── pkg/                        # dkg · kms · signer · storage · network · cron
│   └── BlockchainIndexingService/      # Go: bitcoind RPC, UTXO store, chain state
│       ├── indexers/                   # bitcoin indexer
│       ├── repository/                 # Postgres store (pgx)
│       └── proto/                      # gRPC contracts (address, chain state, transaction)
├── frontend/
│   └── BitcoinWalletFrontend/          # Angular 21 application
├── db/
│   ├── docker-compose.yaml             # PostgreSQL
│   └── create_migration.{sh,bat}       # EF Core migration scripts
├── infrastructure/
│   └── docker-compose.yaml             # Kafka, Schema Registry, AKHQ
└── README.md
```

---

## Implemented Features

- **MPC wallet creation** — distributed key generation (DKG) across MPC parties; job-based orchestration with status polling (`StartKeyGen` / `GetKeyGenStatus` gRPC).
- **Address derivation** — hierarchical (xpub/chain-code based) address derivation from the MPC-generated key.
- **Transactions** — build, sign (threshold ECDSA via MPC `Sign`), and broadcast Bitcoin transactions on **regtest**.
- **Fee estimation** — with in-memory caching.
- **Blockchain indexing** — UTXO tracking, address and chain-state services backed by PostgreSQL; queries bitcoind over JSON-RPC.
- **Service discovery** — full Eureka registration/consumption across all backend services.
- **API Gateway** — Ocelot routing with JWT authentication and RoundRobin load balancing by service name.
- **Database migrations** — EF Core migrations for Wallet and Transaction services via `db/create_migration.sh`.

## Planned / In Progress

- **Frontend** (scaffold only — see [Frontend](#frontend))
- **User & Notification services** — new backend services, not yet started (see [Roadmap](#roadmap))
- **Kafka integration** — connect services with Kafka topics (infrastructure is provisioned)
- **Deployment & CI/CD** — AWS via Terraform (see [Roadmap](#roadmap))
- Additional services (see [Roadmap](#roadmap))

---

## Getting Started

### Prerequisites

- [.NET SDK 8.0](https://dotnet.microsoft.com/download)
- [Go 1.25+](https://go.dev/dl/)
- [JDK 21](https://adoptium.net/)
- [Node.js 20+ / npm 11+](https://nodejs.org/) (frontend)
- [Docker](https://www.docker.com/) + [Docker Compose](https://docs.docker.com/compose/)
- A local `bitcoind` running in **regtest** mode (see [bitcoind](#bitcoind-regtest))

### 1. Start infrastructure

```bash
# PostgreSQL (used by Wallet, Transaction, and Indexing services)
cd db && docker compose up -d

# Kafka, Schema Registry, AKHQ (optional for now — needed once event flows land)
cd infrastructure && docker compose up -d
```

The default DB credentials live in `db/.env`:

```
POSTGRES_USER=admin
POSTGRES_PASSWORD=admin
POSTGRES_DB=database
```

### 2. Apply database migrations

```bash
cd db
./create_migration.sh   # on Windows: create_migration.bat
```

Generates and applies EF Core migrations for `WalletService` and `TransactionService` — see [Database Migrations](#database-migrations) for details on what the scripts do.

### 3. Start the Service Registry

```bash
cd backend/ServiceRegistry
./mvnw spring-boot:run        # Eureka dashboard: http://localhost:8761
```

### 4. Start the backend services

Wait for the registry to come up, then start each service (order matters for registration; the MPC service first, since Wallet/Transaction call it):

```bash
# MPC service (Go)
cd backend/MPC
go run ./cmd/server

# Wallet Service (.NET)
cd backend/NetSolution/WalletService
dotnet run

# Transaction Service (.NET)
cd backend/NetSolution/TransactionService
dotnet run

# Blockchain Indexing Service (Go) — needs bitcoind regtest reachable
cd backend/BlockchainIndexingService
go run .

# API Gateway
cd backend/NetSolution/ApiGateway
dotnet run
```

All services register with Eureka — you can verify them at `http://localhost:8761`.

### 5. bitcoind (regtest)

The Indexing Service expects a regtest node. Either run a local `bitcoind` or use Docker:

```bash
docker run -d --name bitcoind-regtest \
  -p 18443:18443 \
  -v bitcoind-data:/data \
  ruimarinho/bitcoin-core:26 \
  -regtest -server=1 -rpcuser=user -rpcpassword=pass -rpcallowip=0.0.0.0/0 \
  -fallbackfee=0.0001 -txindex=1
```

The connection settings are read from env vars (see [Environment Variables](#environment-variables)).

### 6. Frontend

```bash
cd frontend/BitcoinWalletFrontend
npm install
npm start        # http://localhost:4200
```

> The frontend is an early scaffold — see [Frontend](#frontend) for its current state.

### Smoke test

```bash
# Create a wallet through the gateway
curl -X POST http://localhost:8080/api/v1/wallets

# Wallet status / addresses
curl http://localhost:8080/api/v1/wallets/<wallet-id>
curl http://localhost:8080/api/v1/wallets/<wallet-id>/addresses
```

---

## Database Migrations

`db/create_migration.sh` (Linux/macOS) and `db/create_migration.bat` (Windows) are two variants of the same script. EF Core does not apply migrations on its own, so each run generates a fresh, timestamped migration and applies it to the local database:

1. **Reset the database container** — `docker compose down -v && docker compose up -d` (the `-v` flag deletes the Postgres volume, so **all local data is wiped** — these scripts are for local development only).
2. **WalletService** — `dotnet ef migrations add "Migration_<timestamp>"`, then `dotnet ef database update`.
3. **TransactionService** — the same two commands again.

Migrations are named `Migration_YYYYMMDDHHmm`, so every run produces a unique, ordered migration in the service's `Migrations/` folder.

**Requirements:** the [dotnet-ef CLI tool](https://learn.microsoft.com/ef/core/cli/dotnet-tool) (`dotnet tool install --global dotnet-ef`) and Docker.

The Blockchain Indexing Service does not use EF Core — it creates its own schema on startup (`backend/BlockchainIndexingService/repository/postgres_store.go`).

---

## Environment Variables

| Variable | Default | Used by |
|---|---|---|
| `BTC_RPC_HOST` | `127.0.0.1:18443` | Indexing Service (bitcoind JSON-RPC) |
| `BTC_RPC_USER` | `user` | Indexing Service |
| `BTC_RPC_PASS` | `pass` | Indexing Service |
| `GRPC_PORT` | `:50052` | Indexing Service gRPC listener |
| `PORT` | `:8083` | Indexing Service REST listener |

Connection strings, service addresses and Eureka settings are configured per service in `appsettings.json` (e.g. `MpcService:Address`, `IndexerService:Address`, `ConnectionStrings:WalletDb`).

---

## Ports

| Port | Service |
|---|---|
| `8761` | Service Registry (Eureka) |
| `8080` | API Gateway |
| `8081` / `50053` | Wallet Service (REST / gRPC) |
| `8082` | Transaction Service |
| `8083` / `50052` | Blockchain Indexing Service (REST / gRPC) |
| `50051` | MPC Service (gRPC) |
| `4200` | Frontend (dev server) |
| `5433` | PostgreSQL (host mapping) |
| `9092` | Kafka |
| `3000` | Schema Registry |
| `3001` | AKHQ (Kafka UI) |
| `18443` | bitcoind regtest (JSON-RPC) |

---

## API Overview

All client-facing traffic goes through the API Gateway and is JWT-protected. Routes are defined in `backend/NetSolution/ApiGateway/ocelot.json`.

| Route | Service | Methods |
|---|---|---|
| `/api/v1/wallets` | Wallet-Service | GET, POST, PUT, DELETE, PATCH |
| `/api/v1/wallets/{everything}` | Wallet-Service | GET, POST, PUT, DELETE, PATCH |
| `/api/v1/transactions` | Transaction-Service | GET, POST, PUT, DELETE, PATCH |
| `/api/v1/transactions/{everything}` | Transaction-Service | GET, POST, PUT, DELETE, PATCH |
| `/api/v1/blockchain/{everything}` | Blockchain-Indexing-Service | GET, POST, PUT, DELETE, PATCH |

Inter-service contracts are defined as gRPC protos:

- `backend/MPC/proto/wallet.proto` — `StartKeyGen`, `GetKeyGenStatus`
- `backend/MPC/proto/transaction.proto` — `Sign`
- `backend/BlockchainIndexingService/proto/*.proto` — address, chain state, transaction services

---

## Frontend

Angular 21 application (TypeScript, Angular Router, RxJS) located in `frontend/BitcoinWalletFrontend`.

**Current state:** scaffold only — no wallet features are wired up yet. Planned screens include:

- Wallet dashboard & balance overview
- Receive (addresses) & send flows
- Transaction history
- Key management / backup

<!-- TODO: expand this section as the frontend grows — link to its own README and docs once they exist -->

---

## Roadmap

The sections below are placeholders for upcoming work — they will be filled in as the pieces land.

### New Services

<!-- TODO: add each new service as a row when it is added to the repo -->

| Service | Stack | Purpose | Status |
|---|---|---|---|
| `UserService` | .NET 8 · EF Core · PostgreSQL | user accounts, profiles, auth data | 🚧 Planned |
| `NotificationService` | .NET 8 · Kafka | event-driven notifications (email / push) | 🚧 Planned |
| _(more to come)_ | — | — | — |

### Deployment

**Status: not yet implemented.** Individual services already ship `Dockerfile`s, but there is no CI/CD pipeline or orchestration yet. Target platform: **AWS**, provisioned with **Terraform**. Planned:

- [ ] CI pipeline (build, test, lint) for backend services and frontend
- [ ] Container images for every service (registry)
- [ ] Compose-based environment for staging
- [ ] Kubernetes manifests (deployments, services, ingress)
- [ ] Terraform modules for AWS infrastructure (EKS, RDS, Kafka/MSK, etc.)
- [ ] Secrets management (AWS KMS, DB credentials)
- [ ] Monitoring & logging (metrics, tracing, centralized logs)
- [ ] Analytics dashboards with Apache Superset (Python-based data pipeline)

<!-- TODO: replace the checklist above with concrete docs and links once deployment is implemented -->

### Event-Driven Flows

Kafka + Schema Registry + AKHQ are already provisioned in `infrastructure/docker-compose.yaml` but are not yet consumed by any service. The next step is wiring services together over Kafka topics — wallet events, transaction confirmations, and notifications. The planned `UserService` and `NotificationService` will be the first consumers.

### Production Readiness

- [ ] Mainnet / testnet support (currently regtest-only)
- [ ] Proper KMS integration (hardware-backed key storage for MPC shares)
- [ ] Threat model & security audit documentation
- [ ] Load testing and hardening of the gateway and services

---

## Contributing

<!-- TODO: add contribution guidelines, code style, and PR process once the team settles them -->

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/...`)
3. Commit your changes
4. Open a pull request against `main`

The repository is in active development — features land on the `development` branch and are merged to `main` via pull requests.

---

## License

<!-- TODO: add the license once chosen -->
TBD
