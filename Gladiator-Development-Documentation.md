# GoBook — Complete Development Documentation

> **Codename:** Gladiator  
> **Version:** 2.0  
> **Last Updated:** February 10, 2026  
> **Product Manager:** Claude  
> **Engineering Lead:** Shoaib  
> **Development Tool:** Cursor AI  

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Architecture & System Design](#2-architecture--system-design)
3. [Tech Stack & Dependencies](#3-tech-stack--dependencies)
4. [Monorepo Structure](#4-monorepo-structure)
5. [Development Environment Setup](#5-development-environment-setup)
6. [Database Design with Ent ORM](#6-database-design-with-ent-orm)
7. [Sprint Plan & Feature Stories](#7-sprint-plan--feature-stories)
8. [API Reference](#8-api-reference)
9. [Frontend Architecture](#9-frontend-architecture)
10. [WebSocket Protocol](#10-websocket-protocol)
11. [Error Handling & Logging Standards](#11-error-handling--logging-standards)
12. [Testing Strategy](#12-testing-strategy)
13. [Deployment & DevOps](#13-deployment--devops)
14. [Git Workflow](#14-git-workflow)
15. [Environment Variables](#15-environment-variables)

---

## 1. Executive Summary

**GoBook** is an interactive notebook platform for the Go ecosystem, inspired by Elixir Livebook. It enables developers to write, execute, and share Go code in an interactive browser-based environment with real-time collaboration capabilities.

### Target Users

- Go developers learning new concepts
- Technical educators conducting workshops
- Teams documenting Go patterns and best practices
- Open-source contributors creating interactive tutorials

### Core Value Propositions

- Interactive Go code execution without local toolchain setup
- Personal workspace with cloud persistence
- Collaborative editing for team learning
- Shareable notebooks for knowledge distribution

### Key Architectural Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Repository | Monorepo | Single CI/CD pipeline, simplified deployment |
| Deployment | Single binary on Render | Go backend serves React static files |
| ORM | Ent (Facebook) | Type-safe queries, auto migrations, Go-native |
| Web Framework | Echo | Mature, standard net/http compatible, great WebSocket support |
| Real-time | Gorilla WebSocket | Battle-tested, standard library compatible |
| Frontend | React + TypeScript + Vite | Rich editor support (Monaco), fast builds |
| Local Dev | Docker Compose | Consistent environment across machines |

---

## 2. Architecture & System Design

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Browser                           │
│                    (React SPA :5173)                         │
└──────────────────────────┬──────────────────────────────────┘
                           │
                ┌──────────┴──────────┐
                │                     │
         ┌──────▼──────┐    ┌────────▼────────┐
         │  HTTP/REST  │    │ WebSocket (ws)  │
         │  (JSON)     │    │ (JSON Messages) │
         └──────┬──────┘    └────────┬────────┘
                │                    │
         ┌──────▼────────────────────▼──────────┐
         │     Echo Web Server (Port 8080)      │
         │                                       │
         │  ┌────────────────────────────────┐  │
         │  │  Middleware Stack               │  │
         │  │  - Zerolog Logger              │  │
         │  │  - CORS                        │  │
         │  │  - JWT Auth                    │  │
         │  │  - Rate Limiting               │  │
         │  │  - Request ID                  │  │
         │  │  - Recovery                    │  │
         │  └────────────────────────────────┘  │
         │                                       │
         │  ┌─────────────────────────────────┐ │
         │  │ Handlers (HTTP Routes)          │ │
         │  │  - /api/v1/auth/*              │ │
         │  │  - /api/v1/users/*             │ │
         │  │  - /api/v1/notebooks/*         │ │
         │  │  - /api/v1/notebooks/:id/exec  │ │
         │  │  - /ws/notebooks/:id           │ │
         │  │  - /* (frontend static)        │ │
         │  └──────────────┬──────────────────┘ │
         │                 │                     │
         │  ┌──────────────▼──────────────────┐ │
         │  │ Services (Business Logic)       │ │
         │  │  - AuthService                  │ │
         │  │  - UserService                  │ │
         │  │  - NotebookService              │ │
         │  │  - ShareService                 │ │
         │  │  - ExecutorService              │ │
         │  │  - PermissionService            │ │
         │  └──────────────┬──────────────────┘ │
         │                 │                     │
         │  ┌──────────────▼──────────────────┐ │
         │  │ Data Layer (Ent ORM)            │ │
         │  │  - User entity                  │ │
         │  │  - Notebook entity              │ │
         │  │  - NotebookShare entity         │ │
         │  └──────────────┬──────────────────┘ │
         │                 │                     │
         │  ┌──────────────▼──────────────────┐ │
         │  │ WebSocket Hub                   │ │
         │  │  - Client registry per notebook │ │
         │  │  - Message broadcasting         │ │
         │  │  - Presence tracking            │ │
         │  └─────────────────────────────────┘ │
         └──────┬──────────────────┬─────────────┘
                │                  │
         ┌──────▼──────┐    ┌──────▼─────────┐
         │ PostgreSQL  │    │  Redis         │
         │ (Ent ORM)   │    │  - Sessions    │
         │ - Users     │    │  - Rate limits │
         │ - Notebooks │    │  - Exec cache  │
         │ - Shares    │    │  - Presence    │
         └─────────────┘    └────────────────┘
```

### Data Flow: Code Execution

```
User writes code in Monaco editor cell
         │
         ▼
Frontend debounce-saves cell content (2s)
         │
         ▼
User clicks "Execute" (or Ctrl+Enter)
         │
         ▼
POST /api/v1/notebooks/:id/execute
  body: { cell_id, code }
         │
         ▼
ExecutorService.ExecuteInSession(notebookID, code)
  ├─ Get/Create execution session (Redis)
  ├─ Append code to session context
  ├─ Validate code (block dangerous packages)
  ├─ Execute via yaegi interpreter (or go run)
  └─ Return { status, stdout, stderr, execution_time_ms }
         │
         ▼
Update cell output in PostgreSQL notebook content
         │
         ▼
WebSocket broadcasts cell_output to all connected clients
         │
         ▼
All clients update cell output in real-time
```

### Data Flow: Real-time Collaboration

```
User A edits cell content
         │
         ├─ Optimistic local update (immediate UI feedback)
         │
         ▼
WebSocket sends { type: "cell_update", cell_id, content }
         │
         ▼
Hub broadcasts to all clients in notebook (except sender)
         │
         ▼
User B receives message → updates cell content in editor
```

---

## 3. Tech Stack & Dependencies

### Backend

| Component | Technology | Version | Purpose |
|---|---|---|---|
| Language | Go | 1.24+ | Backend runtime |
| Framework | Echo | v4.11+ | HTTP server, routing, middleware |
| ORM | Ent | v0.13+ | Type-safe database access |
| Database | PostgreSQL | 15+ | Primary data store |
| Cache | Redis | 7+ | Sessions, cache, presence |
| JWT | golang-jwt | v5.2+ | Token-based authentication |
| WebSocket | Gorilla WebSocket | v1.5+ | Real-time communication |
| Interpreter | yaegi | v0.15+ | Go code execution |
| Password Hash | bcrypt | stdlib | Password security |
| Validation | go-playground/validator | v10+ | Input validation |
| Logging | zerolog | v1.31+ | Structured JSON logging |
| Config | godotenv | v1.5+ | Environment variable loading |
| UUID | google/uuid | v1.5+ | Unique identifiers |
| Hot Reload | Air | latest | Development live reload |

### Frontend

| Component | Technology | Version | Purpose |
|---|---|---|---|
| Framework | React | 18.2+ | UI library |
| Language | TypeScript | 5.3+ | Type safety |
| Build Tool | Vite | 5.0+ | Fast bundling & HMR |
| Routing | React Router | 6.21+ | Client-side routing |
| State | Zustand | 4.4+ | Lightweight state management |
| Data Fetching | TanStack React Query | 5.17+ | Server state & caching |
| HTTP Client | Axios | 1.6+ | API requests with interceptors |
| Code Editor | Monaco Editor | 4.6+ | VS Code-powered editor |
| Styling | Tailwind CSS | 3.4+ | Utility-first CSS |
| Icons | Lucide React | 0.303+ | Icon library |
| Markdown | react-markdown | latest | Markdown rendering |

### go.mod Dependencies

```go
module github.com/shoaib/gobook

go 1.24

require (
    entgo.io/ent v0.13.0
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/google/uuid v1.5.0
    github.com/gorilla/websocket v1.5.1
    github.com/joho/godotenv v1.5.1
    github.com/labstack/echo/v4 v4.11.4
    github.com/lib/pq v1.10.9
    github.com/redis/go-redis/v9 v9.4.0
    github.com/rs/zerolog v1.31.0
    github.com/traefik/yaegi v0.15.1
    github.com/go-playground/validator/v10 v10.16.0
    golang.org/x/crypto v0.18.0
)
```

### frontend/package.json Dependencies

```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.21.0",
    "zustand": "^4.4.7",
    "@tanstack/react-query": "^5.17.0",
    "axios": "^1.6.5",
    "@monaco-editor/react": "^4.6.0",
    "tailwindcss": "^3.4.1",
    "lucide-react": "^0.303.0",
    "react-markdown": "^9.0.0",
    "lodash": "^4.17.21"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@types/lodash": "^4.17.0",
    "@vitejs/plugin-react": "^4.2.1",
    "typescript": "^5.3.3",
    "vite": "^5.0.0",
    "autoprefixer": "^10.4.16",
    "postcss": "^8.4.33"
  }
}
```

---

## 4. Monorepo Structure

```
gobook/
├── cmd/
│   └── server/
│       └── main.go                     # Application entry point (Echo server)
├── internal/
│   ├── config/
│   │   └── config.go                   # Configuration loading & validation
│   ├── database/
│   │   ├── ent.go                      # Ent client initialization
│   │   └── redis.go                    # Redis client initialization
│   ├── handlers/
│   │   ├── auth_handler.go             # Auth endpoints
│   │   ├── user_handler.go             # User profile endpoints
│   │   ├── notebook_handler.go         # Notebook CRUD endpoints
│   │   ├── execution_handler.go        # Code execution endpoints
│   │   ├── share_handler.go            # Sharing endpoints
│   │   └── websocket_handler.go        # WebSocket upgrade handler
│   ├── middleware/
│   │   ├── auth.go                     # JWT authentication middleware
│   │   ├── logger.go                   # Zerolog request logger
│   │   └── ratelimit.go               # Redis-based rate limiting
│   ├── services/
│   │   ├── auth_service.go             # Auth business logic
│   │   ├── user_service.go             # User business logic
│   │   ├── notebook_service.go         # Notebook business logic
│   │   ├── share_service.go            # Sharing business logic
│   │   ├── permission_service.go       # Permission checking
│   │   └── executor_service.go         # Code execution service
│   └── websocket/
│       ├── hub.go                      # WebSocket hub (connection registry)
│       ├── client.go                   # WebSocket client (read/write pumps)
│       └── message.go                  # Message type definitions
├── ent/
│   ├── schema/
│   │   ├── user.go                     # User entity schema
│   │   ├── notebook.go                 # Notebook entity schema
│   │   └── notebookshare.go           # NotebookShare entity schema
│   ├── generate.go                     # go:generate directive
│   └── [generated files]              # Auto-generated by Ent
├── pkg/
│   └── executor/
│       ├── executor.go                 # Go code execution engine
│       ├── session.go                  # Execution session management
│       └── validator.go                # Code validation (security)
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── common/                 # ProtectedRoute, LoadingSpinner, ErrorBoundary
│   │   │   ├── auth/                   # LoginForm, RegisterForm
│   │   │   ├── notebook/               # CodeCell, MarkdownCell, CodeEditor, ShareModal, PresenceBar
│   │   │   └── layout/                 # AppLayout, Navbar, Sidebar
│   │   ├── pages/                      # HomePage, LoginPage, RegisterPage, NotebooksPage, NotebookEditorPage, ExplorePage
│   │   ├── hooks/                      # useNotebookWebSocket, useDebounce
│   │   ├── services/                   # api.ts (Axios instance)
│   │   ├── store/                      # authStore.ts, notebookStore.ts (Zustand)
│   │   ├── types/                      # auth.ts, notebook.ts, websocket.ts
│   │   ├── utils/                      # formatters.ts, constants.ts
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── public/
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   └── postcss.config.js
├── docker/
│   ├── Dockerfile                      # Production multi-stage build
│   └── Dockerfile.dev                  # Development with Air hot-reload
├── scripts/
│   └── seed.go                         # Database seeding (optional)
├── .air.toml                           # Air hot-reload configuration
├── docker-compose.yml                  # Local dev stack
├── render.yaml                         # Render IaC blueprint
├── Makefile                            # Common commands
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
└── README.md
```

---

## 5. Development Environment Setup

### Prerequisites

- Docker & Docker Compose
- Go 1.24+
- Node.js 20+
- Make

### Quick Start

```bash
# 1. Clone repository
git clone https://github.com/shoaib/gobook.git && cd gobook

# 2. Copy environment file
cp .env.example .env

# 3. Start all services
make dev

# 4. Access the application
# Backend API:  http://localhost:8080
# Frontend Dev: http://localhost:5173
# PostgreSQL:   localhost:5432
# Redis:        localhost:6379
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: gobook-postgres
    environment:
      POSTGRES_DB: gobook
      POSTGRES_USER: gobook
      POSTGRES_PASSWORD: secret
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U gobook"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - gobook-network

  redis:
    image: redis:7-alpine
    container_name: gobook-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - gobook-network

  backend:
    build:
      context: .
      dockerfile: docker/Dockerfile.dev
    container_name: gobook-backend
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgresql://gobook:secret@postgres:5432/gobook?sslmode=disable
      REDIS_URL: redis://redis:6379
      JWT_SECRET: dev-secret-key-change-in-production
      ENVIRONMENT: development
      PORT: 8080
    volumes:
      - .:/app
      - go_modules:/go/pkg/mod
      - /app/tmp
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - gobook-network
    command: air -c .air.toml

  frontend:
    image: node:20-alpine
    container_name: gobook-frontend
    working_dir: /app/frontend
    ports:
      - "5173:5173"
    environment:
      VITE_API_URL: http://localhost:8080
      VITE_WS_URL: ws://localhost:8080
    volumes:
      - ./frontend:/app/frontend
      - frontend_modules:/app/frontend/node_modules
    command: sh -c "npm install && npm run dev -- --host 0.0.0.0"
    networks:
      - gobook-network

volumes:
  postgres_data:
  redis_data:
  go_modules:
  frontend_modules:

networks:
  gobook-network:
    driver: bridge
```

### docker/Dockerfile.dev

```dockerfile
FROM golang:1.24-alpine
RUN go install github.com/air-verse/air@latest
RUN apk add --no-cache git make
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
EXPOSE 8080
```

### .air.toml

```toml
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["tmp", "vendor", "frontend", "node_modules"]
  exclude_regex = ["_test.go"]
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  stop_on_error = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

### Makefile

```makefile
.PHONY: dev build test generate install-deps lint docker-build

dev:
	docker-compose up

dev-build:
	docker-compose up --build

dev-down:
	docker-compose down

dev-clean:
	docker-compose down -v

install-deps:
	go mod download
	cd frontend && npm install

generate:
	go generate ./ent

build:
	cd frontend && npm run build
	go build -o bin/server ./cmd/server

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

docker-build:
	docker build -f docker/Dockerfile -t gobook:latest .
```

---

## 6. Database Design with Ent ORM

### Entity Relationship Diagram

```
┌─────────────────────────────┐
│           users             │
├─────────────────────────────┤
│ id          UUID (PK)       │
│ email       VARCHAR(255) UQ │
│ password_hash VARCHAR(255)  │
│ name        VARCHAR(255)    │
│ avatar_url  VARCHAR(500)?   │
│ is_active   BOOLEAN         │
│ created_at  TIMESTAMPTZ     │
│ updated_at  TIMESTAMPTZ     │
│ last_login_at TIMESTAMPTZ?  │
└──────────┬──────────────────┘
           │ 1:N (owner_id)
┌──────────▼──────────────────┐
│         notebooks           │
├─────────────────────────────┤
│ id          UUID (PK)       │
│ owner_id    UUID (FK→users) │
│ title       VARCHAR(255)    │
│ description TEXT?            │
│ content     JSONB            │
│ is_public   BOOLEAN          │
│ created_at  TIMESTAMPTZ     │
│ updated_at  TIMESTAMPTZ     │
│ last_executed_at TIMESTAMPTZ?│
│ execution_count INTEGER      │
└──────────┬──────────────────┘
           │ 1:N (cascade delete)
┌──────────▼──────────────────┐
│     notebook_shares         │
├─────────────────────────────┤
│ id          UUID (PK)       │
│ notebook_id UUID (FK→nb)    │
│ shared_with_user_id UUID(FK)│
│ permission  ENUM            │
│   (view | edit | admin)     │
│ shared_by_user_id UUID (FK) │
│ created_at  TIMESTAMPTZ     │
│ UQ(notebook_id,             │
│    shared_with_user_id)     │
└─────────────────────────────┘
```

### JSONB Content Structure (notebooks.content)

```json
{
  "cells": [
    {
      "id": "uuid-string",
      "type": "code",
      "content": "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
      "output": "Hello\n",
      "executed_at": "2026-02-09T10:30:00Z",
      "order": 0
    },
    {
      "id": "uuid-string",
      "type": "markdown",
      "content": "# Section Title\n\nExplanation here...",
      "output": null,
      "executed_at": null,
      "order": 1
    }
  ]
}
```

### Ent Schema: ent/schema/user.go

```go
package schema

import (
    "time"
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
)

type User struct { ent.Schema }

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.String("email").Unique().NotEmpty().MaxLen(255),
        field.String("password_hash").Sensitive().NotEmpty(),
        field.String("name").NotEmpty().MinLen(2).MaxLen(255),
        field.String("avatar_url").Optional().Nillable().MaxLen(500),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
        field.Time("last_login_at").Optional().Nillable(),
        field.Bool("is_active").Default(true),
    }
}

func (User) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("notebooks", Notebook.Type),
        edge.To("shared_notebooks", NotebookShare.Type),
    }
}

func (User) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("email"),
        index.Fields("created_at"),
    }
}
```

### Ent Schema: ent/schema/notebook.go

```go
package schema

import (
    "time"
    "entgo.io/ent"
    "entgo.io/ent/dialect/entsql"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
)

type Notebook struct { ent.Schema }

func (Notebook) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.UUID("owner_id", uuid.UUID{}),
        field.String("title").NotEmpty().MinLen(1).MaxLen(255),
        field.Text("description").Optional().Nillable(),
        field.JSON("content", map[string]interface{}{}).
            Default(map[string]interface{}{"cells": []interface{}{}}),
        field.Bool("is_public").Default(false),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
        field.Time("last_executed_at").Optional().Nillable(),
        field.Int("execution_count").Default(0).NonNegative(),
    }
}

func (Notebook) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("owner", User.Type).Ref("notebooks").
            Field("owner_id").Unique().Required(),
        edge.To("shares", NotebookShare.Type).
            Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
    }
}

func (Notebook) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("owner_id", "created_at"),
        index.Fields("owner_id", "updated_at"),
        index.Fields("is_public", "created_at"),
        index.Fields("updated_at"),
    }
}
```

### Ent Schema: ent/schema/notebookshare.go

```go
package schema

import (
    "time"
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
)

type NotebookShare struct { ent.Schema }

func (NotebookShare) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.UUID("notebook_id", uuid.UUID{}),
        field.UUID("shared_with_user_id", uuid.UUID{}),
        field.Enum("permission").Values("view", "edit", "admin").Default("view"),
        field.UUID("shared_by_user_id", uuid.UUID{}),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

func (NotebookShare) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("notebook", Notebook.Type).Ref("shares").
            Field("notebook_id").Unique().Required(),
        edge.From("shared_with", User.Type).Ref("shared_notebooks").
            Field("shared_with_user_id").Unique().Required(),
    }
}

func (NotebookShare) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("notebook_id", "shared_with_user_id").Unique(),
        index.Fields("notebook_id"),
        index.Fields("shared_with_user_id"),
    }
}
```

### Redis Key Strategy

| Key Pattern | TTL | Purpose |
|---|---|---|
| `session:{user_id}:{token_id}` | 7 days | JWT session tracking |
| `ratelimit:{user_id}:{endpoint}` | 1 minute | Per-user rate limiting |
| `exec_session:{notebook_id}` | 1 hour | Go execution session state |
| `notebook:{notebook_id}` | 15 minutes | Notebook data cache |
| `presence:{notebook_id}:{user_id}` | 5 minutes | User presence in notebook |

---

## 7. Sprint Plan & Feature Stories

### Sprint Overview

| Sprint | Weeks | Focus | Deliverable |
|---|---|---|---|
| Sprint 1 | 1-2 | Project Bootstrap & Infrastructure | Running backend/frontend skeleton with Docker |
| Sprint 2 | 2-3 | Authentication System | Complete auth flow (register, login, JWT, UI) |
| Sprint 3 | 3 | Go Code Execution Engine | Secure code execution with session management |
| Sprint 4 | 4 | Notebook CRUD | Full notebook backend + listing UI |
| Sprint 5 | 5 | Notebook Editor | Monaco editor, cells, code execution UI |
| Sprint 6 | 6-7 | Sharing & Permissions | Multi-user access, sharing UI, public notebooks |
| Sprint 7 | 8 | WebSocket Infrastructure | Real-time server + client connection |
| Sprint 8 | 9-10 | Collaborative Editing | Live editing, presence, shared execution |
| Sprint 9 | 11-12 | Polish & Production | Testing, security, performance, deploy |

---

### Sprint 1: Project Bootstrap & Infrastructure (Weeks 1-2)

**Goal:** Working backend and frontend skeletons with Docker Compose, database connection, and health checks.

---

#### Story 1.1: Initialize Go Module & Project Structure

**Priority:** P0 — Blocker | **Points:** 2 | **Labels:** `backend`, `infrastructure`

**Description:** Create Go module and full monorepo directory structure.

**Acceptance Criteria:**
- [ ] `go mod init github.com/shoaib/gobook` executed
- [ ] All directories from Section 4 created
- [ ] `.gitignore` for Go binaries, Node modules, `.env`, IDE files
- [ ] `.env.example` with all variables from Section 15
- [ ] `README.md` with project description and quick-start
- [ ] `Makefile` with targets: `dev`, `build`, `test`, `generate`, `install-deps`

---

#### Story 1.2: Docker Compose & Local Dev Environment

**Priority:** P0 — Blocker | **Points:** 3 | **Labels:** `infrastructure`, `devops`

**Description:** Docker Compose for local development with PostgreSQL, Redis, backend (Air), frontend (Vite).

**Acceptance Criteria:**
- [ ] `docker-compose.yml` created per Section 5
- [ ] `docker/Dockerfile.dev` with Go + Air
- [ ] `.air.toml` for backend hot-reload
- [ ] `make dev` starts all 4 services successfully
- [ ] Health checks on postgres/redis pass before backend starts
- [ ] `make dev-clean` removes volumes and stops containers

---

#### Story 1.3: Ent ORM Setup & Schema Definition

**Priority:** P0 — Blocker | **Points:** 4 | **Labels:** `backend`, `database`

**Description:** Install Ent, define all 3 entity schemas, generate code, verify auto-migration.

**Acceptance Criteria:**
- [ ] Ent CLI installed
- [ ] `ent/generate.go` with `//go:generate` directive
- [ ] All 3 schemas implemented per Section 6
- [ ] `make generate` runs without errors
- [ ] `internal/database/ent.go` initializes client + runs `client.Schema.Create(ctx)`
- [ ] All indexes and unique constraints verified in PostgreSQL
- [ ] Cascade delete from notebooks to shares verified

---

#### Story 1.4: Redis Client Setup

**Priority:** P0 — Blocker | **Points:** 2 | **Labels:** `backend`, `infrastructure`

**Description:** Initialize Redis client with helpers for sessions, cache, and rate limiting.

**Acceptance Criteria:**
- [ ] `internal/database/redis.go` with connection pool
- [ ] Methods: `SetWithExpiry`, `Get` (JSON unmarshal), `Set` (JSON marshal), `Delete`, `Exists`, `DeleteByPattern`, `Increment`, `Expire`, `Ping`
- [ ] Connection retry with exponential backoff
- [ ] Graceful degradation if Redis unavailable

---

#### Story 1.5: Configuration Management

**Priority:** P0 — Blocker | **Points:** 2 | **Labels:** `backend`, `infrastructure`

**Description:** Centralized config loading with validation and fail-fast.

**Acceptance Criteria:**
- [ ] `internal/config/config.go` with `Config` struct and `Load() (*Config, error)`
- [ ] Loads `.env` in development via `godotenv`
- [ ] Required fields validated: `DATABASE_URL`, `JWT_SECRET`
- [ ] Production guard: rejects dev JWT secret in production
- [ ] Sensible defaults for `PORT`, `ENVIRONMENT`, `REDIS_URL`

---

#### Story 1.6: Echo Server Bootstrap & Health Check

**Priority:** P0 — Blocker | **Points:** 3 | **Labels:** `backend`, `infrastructure`

**Description:** Initialize Echo with middleware stack, health endpoint, graceful shutdown.

**Acceptance Criteria:**
- [ ] `cmd/server/main.go` as entry point
- [ ] Middleware: Recovery, RequestID, CORS (localhost:5173 + production)
- [ ] `GET /health` returns 200 with `{"status": "healthy"}` or 503 with error
- [ ] Graceful shutdown on SIGINT/SIGTERM (10s timeout)
- [ ] Startup log: `Server started on port 8080`

---

#### Story 1.7: Zerolog Request Logger Middleware

**Priority:** P1 — High | **Points:** 2 | **Labels:** `backend`, `observability`

**Description:** Replace Echo's default logger with zerolog structured logging.

**Acceptance Criteria:**
- [ ] `internal/middleware/logger.go` — Echo middleware
- [ ] Fields: method, path, status, latency, ip, user_agent, request_id
- [ ] Development: console writer (human-readable)
- [ ] Production: JSON format

---

#### Story 1.8: React Frontend Initialization

**Priority:** P0 — Blocker | **Points:** 3 | **Labels:** `frontend`, `infrastructure`

**Description:** Initialize React + TypeScript + Vite + Tailwind with routing and API client.

**Acceptance Criteria:**
- [ ] Vite React-TS template in `frontend/`
- [ ] Tailwind CSS configured
- [ ] React Router with all routes: `/`, `/login`, `/register`, `/notebooks`, `/notebooks/:id`, `/explore`
- [ ] `src/services/api.ts` — Axios instance with auth interceptor (Bearer token) and 401 redirect
- [ ] `src/components/common/ProtectedRoute.tsx`
- [ ] `src/types/` — TypeScript interfaces for User, Notebook, Cell, TokenPair
- [ ] Environment: `VITE_API_URL`, `VITE_WS_URL`

---

#### Story 1.9: Production Dockerfile (Multi-stage)

**Priority:** P0 — Blocker | **Points:** 3 | **Labels:** `devops`, `deployment`

**Description:** Multi-stage Dockerfile: build frontend, build backend, serve as single binary.

**Acceptance Criteria:**
- [ ] 3 stages: frontend-builder (Node), backend-builder (Go), final (Alpine)
- [ ] Backend serves frontend via Go `embed` directive
- [ ] SPA routing: `index.html` for all non-API routes
- [ ] Non-root user, health check configured
- [ ] Final image < 100MB
- [ ] `make docker-build` succeeds

---

#### Story 1.10: Render Deployment Configuration

**Priority:** P1 — High | **Points:** 2 | **Labels:** `devops`, `deployment`

**Description:** Render IaC blueprint for PostgreSQL + Redis + web service.

**Acceptance Criteria:**
- [ ] `render.yaml` with database, Redis, web service
- [ ] Environment variables from Render services
- [ ] Auto-deploy from `main` branch
- [ ] Health check: `/health`

---

### Sprint 2: Authentication System (Week 2-3)

**Goal:** Complete auth flow — register, login, JWT tokens, refresh, logout, frontend auth pages.

---

#### Story 2.1: User Registration Backend

**Priority:** P0 — Critical | **Points:** 3 | **Labels:** `backend`, `auth`

**Description:** `POST /api/v1/auth/register` with validation, bcrypt, Ent user creation.

**Acceptance Criteria:**
- [ ] `internal/services/auth_service.go` — `Register(ctx, req) (*ent.User, error)`
- [ ] `internal/handlers/auth_handler.go` — `Register(c echo.Context) error`
- [ ] Validation: email (required, valid, max 255), password (min 8, max 72), name (min 2, max 255)
- [ ] Bcrypt cost factor 12
- [ ] Duplicate email → 409 Conflict
- [ ] Success → 201 Created with user (password excluded via `Sensitive()`)

**Request/Response:**
```
POST /api/v1/auth/register
{ "email": "shoaib@example.com", "password": "SecurePass123", "name": "Shoaib" }

→ 201 { "user": { "id": "uuid", "email": "...", "name": "...", "created_at": "..." } }
```

---

#### Story 2.2: JWT Login & Token Generation

**Priority:** P0 — Critical | **Points:** 4 | **Labels:** `backend`, `auth`

**Description:** `POST /api/v1/auth/login` with JWT access/refresh tokens + Redis session.

**Acceptance Criteria:**
- [ ] Ent user lookup: `client.User.Query().Where(user.EmailEQ(email), user.IsActiveEQ(true)).Only(ctx)`
- [ ] bcrypt password comparison
- [ ] Access token: HS256, 15min, claims: `user_id`, `email`, `token_id`, `exp`, `iat`
- [ ] Refresh token: HS256, 7 days, claims: `user_id`, `token_id`, `type: "refresh"`
- [ ] Redis session: `session:{user_id}:{token_id}`, TTL 7 days
- [ ] Updates `last_login_at`
- [ ] Invalid credentials → 401

**Response:**
```json
{ "access_token": "...", "refresh_token": "...", "expires_in": 900, "user": { ... } }
```

---

#### Story 2.3: JWT Auth Middleware

**Priority:** P0 — Critical | **Points:** 3 | **Labels:** `backend`, `auth`, `middleware`

**Description:** Echo middleware validating JWT + Redis session, injecting user_id into context.

**Acceptance Criteria:**
- [ ] `internal/middleware/auth.go` — `JWTAuth(authService) echo.MiddlewareFunc`
- [ ] Extracts from `Authorization: Bearer <token>`
- [ ] Validates: signature (HS256), expiry, Redis session
- [ ] Sets `c.Set("user_id", userID)` and `c.Set("token_id", tokenID)`
- [ ] Returns 401 JSON for all failure cases

---

#### Story 2.4: Token Refresh & Logout

**Priority:** P0 — Critical | **Points:** 2 | **Labels:** `backend`, `auth`

**Description:** Token refresh for session extension, logout for termination.

**Acceptance Criteria:**
- [ ] `POST /api/v1/auth/refresh` — validates refresh token, returns new access token (same token_id)
- [ ] `POST /api/v1/auth/logout` (protected) — deletes Redis session → 200
- [ ] `POST /api/v1/auth/logout-all` (protected) — deletes all sessions → 200

---

#### Story 2.5: Rate Limiting Middleware

**Priority:** P1 — High | **Points:** 2 | **Labels:** `backend`, `middleware`, `security`

**Description:** Redis-based rate limiting per endpoint group.

**Acceptance Criteria:**
- [ ] Auth: 5/min per IP, API: 100/min per user, Execution: 10/min per notebook
- [ ] Redis INCR + EXPIRE (sliding window)
- [ ] Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`
- [ ] 429 when exceeded

---

#### Story 2.6: Auth UI — Login & Register Pages

**Priority:** P0 — Critical | **Points:** 5 | **Labels:** `frontend`, `auth`

**Description:** React login/register pages with Zustand auth store.

**Acceptance Criteria:**
- [ ] `src/store/authStore.ts` (Zustand + persist):
  - State: `user`, `accessToken`, `refreshToken`, `isAuthenticated`, `isLoading`
  - Actions: `login()`, `register()`, `logout()`, `refreshAccessToken()`
  - Auto-refresh timer at 13-minute mark
- [ ] `src/pages/LoginPage.tsx`: email, password (toggle visibility), submit, error display, link to register
- [ ] `src/pages/RegisterPage.tsx`: email, name, password, confirm password, strength indicator, auto-login after success
- [ ] `src/components/common/ProtectedRoute.tsx`: check auth, loading spinner, redirect to /login
- [ ] Responsive card layout, Axios interceptor clears auth on 401

---

### Sprint 3: Go Code Execution Engine (Week 3)

**Goal:** Secure Go code execution with session persistence.

---

#### Story 3.1: Execution Engine — Interpreted Mode (yaegi)

**Priority:** P0 — Critical | **Points:** 5 | **Labels:** `backend`, `execution`

**Description:** Core Go code execution engine using yaegi with security constraints.

**Acceptance Criteria:**
- [ ] `pkg/executor/executor.go` — `Executor.Execute(ctx, req) (*ExecutionResult, error)`
- [ ] Interpreted mode (yaegi): `interp.New` with captured stdout/stderr, `i.Use(stdlib.Symbols)`, `i.Eval(code)`
- [ ] Compiled mode (fallback): temp file + `go run` via `exec.CommandContext`
- [ ] Timeout: 30s max via `context.WithTimeout`
- [ ] Forbidden packages blocked: `os/exec`, `syscall`, `unsafe`
- [ ] `ExecutionResult`: `{ status, stdout, stderr, execution_time_ms, compiled, error }`
- [ ] Timeout returns `status: "timeout"`, errors return `status: "error"` with details

---

#### Story 3.2: Execution Session Management

**Priority:** P1 — High | **Points:** 3 | **Labels:** `backend`, `execution`

**Description:** Redis-based execution sessions for state persistence across cells.

**Acceptance Criteria:**
- [ ] `pkg/executor/session.go` — `SessionManager`
- [ ] `GetOrCreateSession(notebookID)` — from Redis or new
- [ ] `ExecuteInSession(notebookID, code)` — accumulates code, executes full context
- [ ] `ClearSession(notebookID)` — deletes from Redis
- [ ] Redis key: `exec_session:{notebook_id}`, TTL 1 hour
- [ ] Failed code NOT appended to session

---

#### Story 3.3: Execution REST Endpoint

**Priority:** P0 — Critical | **Points:** 2 | **Labels:** `backend`, `execution`, `api`

**Description:** HTTP endpoint for cell execution with access checks.

**Acceptance Criteria:**
- [ ] `POST /api/v1/notebooks/:id/execute` (protected) — `{ "cell_id": "...", "code": "..." }`
- [ ] Validates user access (owner or editor)
- [ ] Updates cell output in notebook JSONB content
- [ ] Returns ExecutionResult
- [ ] `GET /api/v1/notebooks/:id/session` — session info
- [ ] `DELETE /api/v1/notebooks/:id/session` — clear session

---

### Sprint 4: Notebook CRUD Backend & Frontend (Week 4)

**Goal:** Complete notebook CRUD with API and listing UI.

---

#### Story 4.1: Create Notebook Backend

**Priority:** P0 — Critical | **Points:** 3 | **Labels:** `backend`, `notebook`, `api`

**Description:** `POST /api/v1/notebooks` with Ent, default cells, owner from JWT.

**Acceptance Criteria:**
- [ ] Request: `{ "title": "...", "description": "...", "initial_cells": [...] }`
- [ ] Default cells if empty: markdown + code cell
- [ ] Ent: `client.Notebook.Create().SetOwnerID(uid).SetTitle(...).SetContent(...).Save(ctx)`
- [ ] Returns 201 with full notebook
- [ ] Validates: title 1-255 chars, description max 1000

---

#### Story 4.2: List Notebooks Backend

**Priority:** P0 — Critical | **Points:** 3 | **Labels:** `backend`, `notebook`, `api`

**Description:** `GET /api/v1/notebooks` with pagination, search, sorting via Ent.

**Acceptance Criteria:**
- [ ] Query params: `page`, `limit`, `sort`, `order`, `search`
- [ ] Returns only user's owned notebooks (Phase 1)
- [ ] Search: `TitleContainsFold` OR `DescriptionContainsFold`
- [ ] Metadata only (no full content): id, title, description, is_public, timestamps, cell_count
- [ ] Pagination: `{ notebooks: [...], pagination: { page, limit, total, total_pages } }`

---

#### Story 4.3: Get Single Notebook Backend

**Priority:** P0 — Critical | **Points:** 3 | **Labels:** `backend`, `notebook`, `api`

**Description:** `GET /api/v1/notebooks/:id` with eager loading and permission check.

**Acceptance Criteria:**
- [ ] Ent: `client.Notebook.Query().Where(IDEQ(nid)).WithOwner().Only(ctx)`
- [ ] Returns full content, owner profile, computed permissions
- [ ] 404 if not found, 403 if no access
- [ ] Redis cache (15min TTL), invalidated on update/delete

---

#### Story 4.4: Update Notebook Backend (Optimistic Locking)

**Priority:** P0 — Critical | **Points:** 3 | **Labels:** `backend`, `notebook`, `api`

**Description:** `PATCH /api/v1/notebooks/:id` with transaction and `updated_at` check.

**Acceptance Criteria:**
- [ ] Partial updates (only provided fields)
- [ ] Transaction: `ForUpdate()` → compare `updated_at` → apply or 409 Conflict
- [ ] Invalidates Redis cache
- [ ] Returns updated notebook

---

#### Story 4.5: Delete Notebook Backend

**Priority:** P1 — High | **Points:** 2 | **Labels:** `backend`, `notebook`, `api`

**Description:** `DELETE /api/v1/notebooks/:id` with ownership check.

**Acceptance Criteria:**
- [ ] Owner only (403 otherwise)
- [ ] Cascade deletes shares (Ent schema config)
- [ ] Cleans up Redis (cache + execution session)
- [ ] Returns 204 No Content

---

#### Story 4.6: Notebooks List Page (Frontend)

**Priority:** P0 — Critical | **Points:** 5 | **Labels:** `frontend`, `notebook`

**Description:** Grid/list view with search, sort, pagination for user's notebooks.

**Acceptance Criteria:**
- [ ] `src/pages/NotebooksPage.tsx` at `/notebooks`
- [ ] React Query for data fetching
- [ ] Grid/list toggle, search (debounced), sort dropdown
- [ ] "New Notebook" button
- [ ] Card: title, description, cell count, last updated, public/private badge
- [ ] Click → `/notebooks/:id`, hover → Edit/Delete actions
- [ ] Delete confirmation dialog
- [ ] Pagination, empty state, loading skeleton

---

### Sprint 5: Notebook Editor (Week 5)

**Goal:** Full notebook editing with Monaco editor, markdown cells, and code execution.

---

#### Story 5.1: Monaco Editor Component

**Priority:** P0 — Blocker | **Points:** 4 | **Labels:** `frontend`, `editor`

**Description:** Reusable Monaco wrapper with Go syntax highlighting.

**Acceptance Criteria:**
- [ ] `src/components/notebook/CodeEditor.tsx`
- [ ] Go language syntax highlighting
- [ ] Auto-grows (100px–600px)
- [ ] `Ctrl+Enter` → `onExecute`, `Ctrl+S` → `onSave`
- [ ] No minimap, line numbers, font 14, tab 4, word wrap
- [ ] Props: `value`, `onChange`, `onExecute`, `readOnly`, `theme`

---

#### Story 5.2: Code Cell Component

**Priority:** P0 — Critical | **Points:** 5 | **Labels:** `frontend`, `editor`

**Description:** Code cell with editor, execute button, output display.

**Acceptance Criteria:**
- [ ] `src/components/notebook/CodeCell.tsx`
- [ ] Toolbar: Play (execute), time badge, + Cell, Delete
- [ ] Execute → POST, loading spinner, output display
- [ ] Output: collapsible, success (gray), error (red), execution time
- [ ] Selection highlight (blue ring)

---

#### Story 5.3: Markdown Cell Component

**Priority:** P0 — Critical | **Points:** 3 | **Labels:** `frontend`, `editor`

**Description:** Markdown cell with edit/view toggle.

**Acceptance Criteria:**
- [ ] `src/components/notebook/MarkdownCell.tsx`
- [ ] Edit: textarea, View: rendered via `react-markdown`
- [ ] Toggle button, double-click to edit, blur to save
- [ ] Delete button

---

#### Story 5.4: Notebook Editor Page

**Priority:** P0 — Critical | **Points:** 5 | **Labels:** `frontend`, `editor`

**Description:** Main editor page orchestrating cells, state, and backend sync.

**Acceptance Criteria:**
- [ ] `src/pages/NotebookEditorPage.tsx` at `/notebooks/:id`
- [ ] Fetch notebook on mount, render cells in order
- [ ] Inline title/description editing
- [ ] Save status: "Saving..." / "All changes saved" / "Save failed"
- [ ] Auto-save (2s debounce)
- [ ] Add/delete/update cells, cell type picker
- [ ] Optimistic UI, responsive layout

---

### Sprint 6: Sharing & Permissions (Weeks 6-7)

**Goal:** Multi-user sharing with permissions and public notebooks.

---

#### Story 6.1: Share Backend — Endpoints & Service

**Priority:** P0 — Critical | **Points:** 4 | **Labels:** `backend`, `sharing`

**Description:** Share notebooks with users by email via Ent.

**Acceptance Criteria:**
- [ ] `POST /api/v1/notebooks/:id/share` — `{ "email": "...", "permission": "view|edit|admin" }`, owner only
- [ ] `GET /api/v1/notebooks/:id/shares` — list shares (owner only)
- [ ] `PATCH /api/v1/notebooks/:id/shares/:sid` — update permission
- [ ] `DELETE /api/v1/notebooks/:id/shares/:sid` — revoke share
- [ ] Validation: user exists, not self, no duplicates

---

#### Story 6.2: Permission Service

**Priority:** P0 — Critical | **Points:** 3 | **Labels:** `backend`, `sharing`

**Description:** Centralized permission checking considering ownership + shares.

**Acceptance Criteria:**
- [ ] `internal/services/permission_service.go`
- [ ] `CheckNotebookAccess(ctx, nbID, userID) → { CanView, CanEdit, CanDelete, IsOwner }`
- [ ] Owner → all true; admin/edit share → view+edit; view share → view only
- [ ] All notebook endpoints updated to use PermissionService

---

#### Story 6.3: Shared Notebooks in Listing

**Priority:** P1 — High | **Points:** 2 | **Labels:** `backend`, `sharing`

**Description:** Include shared notebooks in user's listing.

**Acceptance Criteria:**
- [ ] `GET /api/v1/notebooks` returns owned + shared notebooks
- [ ] Each includes `role`: "owner" | "editor" | "viewer"

---

#### Story 6.4: Share Modal (Frontend)

**Priority:** P1 — High | **Points:** 5 | **Labels:** `frontend`, `sharing`

**Description:** Sharing modal UI with email input and permission management.

**Acceptance Criteria:**
- [ ] `src/components/notebook/ShareModal.tsx`
- [ ] Email + permission dropdown + Share button
- [ ] Current shares list: avatar, name, permission badge, revoke button
- [ ] Error handling, React Query mutations, toast notifications

---

#### Story 6.5: Public Notebooks & Fork

**Priority:** P1 — High | **Points:** 3 | **Labels:** `backend`, `frontend`, `sharing`

**Acceptance Criteria:**
- [ ] `GET /api/v1/notebooks/public` — paginated, searchable (no auth)
- [ ] `POST /api/v1/notebooks/:id/fork` — creates personal copy
- [ ] `src/pages/ExplorePage.tsx` — grid of public notebooks with Fork button

---

### Sprint 7: WebSocket & Real-time Infrastructure (Week 8)

**Goal:** WebSocket server with connection management and message protocol.

---

#### Story 7.1: WebSocket Hub & Client

**Priority:** P0 — Blocker | **Points:** 5 | **Labels:** `backend`, `websocket`

**Description:** WebSocket hub managing connections per notebook with broadcasting.

**Acceptance Criteria:**
- [ ] `internal/websocket/hub.go` — Hub with notebooks map, register/unregister/broadcast channels, `Run()` goroutine
- [ ] `internal/websocket/client.go` — Client with `ReadPump()`, `WritePump()`, ping/pong (30s), read deadline (60s)
- [ ] `internal/websocket/message.go` — Types: `user_joined`, `user_left`, `cell_update`, `cell_execute`, `cell_output`, `cursor_move`, `cell_lock`, `cell_unlock`
- [ ] `GetActiveUsers(notebookID)` returns current users

---

#### Story 7.2: WebSocket Handler with Auth

**Priority:** P0 — Critical | **Points:** 3 | **Labels:** `backend`, `websocket`, `auth`

**Description:** Echo handler upgrading HTTP → WebSocket with JWT auth.

**Acceptance Criteria:**
- [ ] `GET /ws/notebooks/:id?token=JWT`
- [ ] Validates JWT from query param + Redis session
- [ ] Checks notebook access via PermissionService
- [ ] Upgrades, creates Client, registers with Hub
- [ ] On disconnect: unregisters, broadcasts `user_left`

---

#### Story 7.3: Frontend WebSocket Hook

**Priority:** P0 — Critical | **Points:** 4 | **Labels:** `frontend`, `websocket`

**Description:** React hook with auto-reconnect and message handling.

**Acceptance Criteria:**
- [ ] `src/hooks/useNotebookWebSocket.ts`
- [ ] Returns: `{ isConnected, activeUsers, sendMessage }`
- [ ] Auto-reconnect with exponential backoff (1s → 30s max)
- [ ] Tracks activeUsers from join/leave messages
- [ ] Cleanup on unmount

---

### Sprint 8: Collaborative Editing (Weeks 9-10)

**Goal:** Real-time cell sync, presence, shared execution.

---

#### Story 8.1: User Presence Indicators

**Priority:** P1 — High | **Points:** 3 | **Labels:** `frontend`, `collaboration`

**Description:** Active user avatars in notebook editor header.

**Acceptance Criteria:**
- [ ] `src/components/notebook/PresenceBar.tsx`
- [ ] Color-coded avatar circles (first letter), max 5 + overflow
- [ ] "You" tooltip for current user
- [ ] Real-time updates, count display

---

#### Story 8.2: Real-time Cell Updates

**Priority:** P0 — Critical | **Points:** 5 | **Labels:** `frontend`, `collaboration`

**Description:** Sync cell content via WebSocket with cell locking.

**Acceptance Criteria:**
- [ ] On edit: debounced WebSocket send (500ms) of `cell_update`
- [ ] On receive: apply content (skip self)
- [ ] Cell locking: `cell_lock` on focus, `cell_unlock` on blur
- [ ] "User X is editing" banner on locked cells (read-only for others)
- [ ] Optimistic local updates

---

#### Story 8.3: Shared Execution Results

**Priority:** P1 — High | **Points:** 3 | **Labels:** `backend`, `frontend`, `collaboration`

**Description:** Broadcast execution results to all connected clients.

**Acceptance Criteria:**
- [ ] Backend broadcasts `cell_execute` (start) and `cell_output` (result) via Hub
- [ ] "User X is running" indicator
- [ ] Output updates for all users
- [ ] Execution badge shows who last ran cell

---

### Sprint 9: Polish, Testing & Production Readiness (Weeks 11-12)

**Goal:** Performance, testing, security, deployment.

---

#### Story 9.1: Database Query Optimization

**Priority:** P1 — High | **Points:** 3 | **Labels:** `backend`, `performance`

**Acceptance Criteria:**
- [ ] All Ent schema indexes verified in PostgreSQL
- [ ] Redis caching: notebook metadata (15min), user shares (5min)
- [ ] Target: list < 50ms, get < 20ms

---

#### Story 9.2: Frontend Performance

**Priority:** P1 — High | **Points:** 4 | **Labels:** `frontend`, `performance`

**Acceptance Criteria:**
- [ ] Route-level code splitting (`React.lazy`)
- [ ] Monaco Editor lazy loaded
- [ ] `React.memo` on cell components
- [ ] Bundle < 500KB gzipped, Lighthouse > 90

---

#### Story 9.3: Security Hardening

**Priority:** P0 — Critical | **Points:** 4 | **Labels:** `backend`, `security`

**Acceptance Criteria:**
- [ ] CORS whitelist (production + localhost)
- [ ] Security headers: CSP, X-Frame-Options, HSTS
- [ ] Execution sandboxing: forbidden packages, timeout enforcement
- [ ] Input validation on ALL endpoints
- [ ] JWT algorithm rejection for non-HS256

---

#### Story 9.4: Keyboard Shortcuts

**Priority:** P1 — High | **Points:** 2 | **Labels:** `frontend`, `ux`

**Acceptance Criteria:**
- [ ] `?` → help modal
- [ ] `↑/↓` → cell navigation, `Enter` → edit, `Esc` → exit
- [ ] `Ctrl+Enter` → execute, `Shift+Enter` → execute + next
- [ ] `B` → cell below, `A` → cell above, `M` → markdown, `Y` → code
- [ ] `Ctrl+S` → save

---

#### Story 9.5: Unit & Integration Tests

**Priority:** P0 — Critical | **Points:** 8 | **Labels:** `testing`

**Acceptance Criteria:**
- [ ] Backend: auth, notebook, executor, share service tests (70%+ coverage)
- [ ] Handler integration tests with httptest
- [ ] Frontend: component tests (React Testing Library) for auth forms, notebook list, code cell
- [ ] CI runs tests on PR

---

#### Story 9.6: Docker Build & Render Deployment

**Priority:** P0 — Critical | **Points:** 4 | **Labels:** `devops`, `deployment`

**Acceptance Criteria:**
- [ ] Production Dockerfile builds and runs
- [ ] `render.yaml` provisions all resources
- [ ] Auto-deploy from `main` verified
- [ ] HTTPS working, smoke test passes: register → login → create → execute → share

---

#### Story 9.7: Export Notebook (Nice to have)

**Priority:** P2 | **Points:** 3 | **Labels:** `backend`, `feature`

**Acceptance Criteria:**
- [ ] `GET /api/v1/notebooks/:id/export?format=gobook|md|go`
- [ ] `.gobook` (JSON), `.md` (markdown), `.go` (standalone file)
- [ ] Download button in editor settings

---

#### Story 9.8: Notebook Templates (Nice to have)

**Priority:** P2 | **Points:** 3 | **Labels:** `feature`

**Acceptance Criteria:**
- [ ] System-seeded template notebooks (public)
- [ ] Templates: "Go Concurrency", "HTTP Servers", "Testing"
- [ ] "Create from template" option → forks template

---

## 8. API Reference

### Authentication

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/auth/register` | No | Register new user |
| POST | `/api/v1/auth/login` | No | Login, get tokens |
| POST | `/api/v1/auth/refresh` | No | Refresh access token |
| POST | `/api/v1/auth/logout` | Yes | Logout current session |
| POST | `/api/v1/auth/logout-all` | Yes | Logout all sessions |

### Users

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/users/me` | Yes | Get current user profile |
| PATCH | `/api/v1/users/me` | Yes | Update profile (name, avatar) |

### Notebooks

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/notebooks` | Yes | List user's notebooks |
| POST | `/api/v1/notebooks` | Yes | Create new notebook |
| GET | `/api/v1/notebooks/:id` | Yes | Get notebook with full content |
| PATCH | `/api/v1/notebooks/:id` | Yes | Update notebook (optimistic lock) |
| DELETE | `/api/v1/notebooks/:id` | Yes | Delete notebook (owner only) |
| GET | `/api/v1/notebooks/public` | No | List public notebooks |
| POST | `/api/v1/notebooks/:id/fork` | Yes | Fork public notebook |
| GET | `/api/v1/notebooks/:id/export` | Yes | Export notebook |

### Execution

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/notebooks/:id/execute` | Yes | Execute code cell |
| GET | `/api/v1/notebooks/:id/session` | Yes | Get execution session |
| DELETE | `/api/v1/notebooks/:id/session` | Yes | Clear execution session |

### Sharing

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/notebooks/:id/share` | Yes | Share notebook |
| GET | `/api/v1/notebooks/:id/shares` | Yes | List shares (owner) |
| PATCH | `/api/v1/notebooks/:id/shares/:sid` | Yes | Update permission |
| DELETE | `/api/v1/notebooks/:id/shares/:sid` | Yes | Revoke share |

### WebSocket

| Path | Auth | Description |
|---|---|---|
| `ws://host/ws/notebooks/:id?token=JWT` | Query param | Real-time collaboration |

### Standard Error Response

```json
{ "error": "error_code_or_message" }
```

### HTTP Status Codes

| Code | Meaning |
|---|---|
| 200 | Success |
| 201 | Created |
| 204 | No Content (delete) |
| 400 | Bad Request (validation) |
| 401 | Unauthorized |
| 403 | Forbidden (permission) |
| 404 | Not Found |
| 409 | Conflict (duplicate / optimistic lock) |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

---

## 9. Frontend Architecture

### State Management

- **Zustand** (`authStore`): user, tokens, auth status, login/logout/refresh
- **React Query**: server state (notebooks, shares, execution results)

### Route Structure

```typescript
<Routes>
  <Route path="/" element={<HomePage />} />
  <Route path="/login" element={<LoginPage />} />
  <Route path="/register" element={<RegisterPage />} />
  <Route path="/explore" element={<ExplorePage />} />
  <Route element={<ProtectedRoute />}>
    <Route path="/notebooks" element={<NotebooksPage />} />
    <Route path="/notebooks/:id" element={<NotebookEditorPage />} />
  </Route>
</Routes>
```

### Key TypeScript Interfaces

```typescript
interface Cell {
  id: string;
  type: 'code' | 'markdown';
  content: string;
  output: string | null;
  executed_at: string | null;
  order: number;
}

interface Notebook {
  id: string;
  owner_id: string;
  title: string;
  description: string | null;
  content: { cells: Cell[] };
  is_public: boolean;
  created_at: string;
  updated_at: string;
  last_executed_at: string | null;
  execution_count: number;
}

interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string | null;
}

interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

interface WebSocketMessage {
  type: string;
  notebook_id: string;
  user_id: string;
  user_name: string;
  payload: any;
  timestamp: number;
}
```

---

## 10. WebSocket Protocol

### Message Types

| Type | Direction | Payload | Description |
|---|---|---|---|
| `user_joined` | Server → Client | `{ user_id, user_name }` | User connected |
| `user_left` | Server → Client | `{ user_id, user_name }` | User disconnected |
| `cell_update` | Bidirectional | `{ cell_id, content }` | Cell content changed |
| `cell_lock` | Bidirectional | `{ cell_id, user_name }` | User editing cell |
| `cell_unlock` | Bidirectional | `{ cell_id }` | User stopped editing |
| `cell_execute` | Server → Client | `{ cell_id, user_name }` | Execution started |
| `cell_output` | Server → Client | `{ cell_id, output, error, status }` | Execution completed |
| `cursor_move` | Bidirectional | `{ cell_id, position }` | Cursor position |

### Connection Lifecycle

```
Client                              Server
  │  GET /ws/notebooks/:id?token=JWT   │
  │ ──────────────────────────────────>│ Validate JWT + access
  │  101 Switching Protocols           │
  │ <──────────────────────────────────│ Register client
  │  ← { type: "user_joined" }        │ Broadcast
  │  → { type: "cell_update" }        │
  │  ← { type: "cell_update" }        │ Broadcast to others
  │  ← PING (every 30s)               │
  │  → PONG                           │
  │  [disconnect]                      │ Unregister
  │  ← { type: "user_left" }          │ Broadcast
```

---

## 11. Error Handling & Logging Standards

### Backend Error Pattern

```go
// Sentinel errors in services
var (
    ErrNotFound           = errors.New("not found")
    ErrAccessDenied       = errors.New("access denied")
    ErrConflict           = errors.New("conflict")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrEmailExists        = errors.New("email already exists")
    ErrRateLimitExceeded  = errors.New("rate limit exceeded")
)

// Handlers map errors → HTTP status
// ErrNotFound → 404, ErrAccessDenied → 403, ErrConflict → 409, etc.
```

### Logging Standards

```go
// INFO: successful operations
log.Info().Str("user_id", uid).Str("notebook_id", nbID).Msg("notebook created")

// WARN: recoverable issues
log.Warn().Err(err).Msg("redis unavailable, continuing without cache")

// ERROR: failures
log.Error().Err(err).Str("user_id", uid).Msg("failed to create notebook")

// DEBUG: development details
log.Debug().Str("query", q).Int("results", count).Msg("notebook search")
```

---

## 12. Testing Strategy

### Backend Tests

| Layer | Tool | Coverage Target | What to Test |
|---|---|---|---|
| Services | `go test` | 80% | Business logic, Ent queries, errors |
| Handlers | `httptest` | 70% | Request parsing, status codes, response |
| Executor | `go test` | 90% | Execution, timeout, security blocking |
| Middleware | `go test` | 70% | Auth validation, rate limiting |

### Frontend Tests

| Layer | Tool | Coverage Target | What to Test |
|---|---|---|---|
| Components | React Testing Library | 60% | Rendering, interactions, state |
| Hooks | `renderHook` | 70% | WebSocket, auth store |
| E2E | Playwright (future) | Critical paths | Full user flows |

### Running Tests

```bash
make test                               # All Go tests
go test -v ./internal/services/...      # Specific package
go test -cover ./...                    # Coverage
cd frontend && npm test                 # Frontend tests
```

---

## 13. Deployment & DevOps

### Production Dockerfile

```dockerfile
# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --only=production
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.24-alpine AS backend-builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN go generate ./ent
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Stage 3: Final
FROM alpine:latest
RUN apk --no-cache add ca-certificates go
WORKDIR /app
COPY --from=backend-builder /app/server .
COPY --from=backend-builder /app/frontend/dist ./frontend/dist
RUN addgroup -g 1001 -S gobook && adduser -u 1001 -S gobook -G gobook
RUN chown -R gobook:gobook /app
USER gobook
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD wget --spider http://localhost:8080/health || exit 1
CMD ["./server"]
```

### render.yaml

```yaml
services:
  - type: web
    name: gobook
    env: docker
    region: oregon
    plan: starter
    dockerfilePath: ./docker/Dockerfile
    dockerContext: .
    branch: main
    healthCheckPath: /health
    envVars:
      - key: ENVIRONMENT
        value: production
      - key: PORT
        value: 8080
      - key: DATABASE_URL
        fromDatabase:
          name: gobook-db
          property: connectionString
      - key: REDIS_URL
        fromService:
          type: redis
          name: gobook-redis
          property: connectionString
      - key: JWT_SECRET
        generateValue: true
      - key: APP_URL
        value: https://gobook.onrender.com
    autoDeploy: true

databases:
  - name: gobook-db
    plan: starter
    databaseName: gobook
    user: gobook
```

### Deployment Flow

```
Push to main → Render detects → Build Docker → Health check → Auto-migrate (Ent) → Deploy
```

---

## 14. Git Workflow

### Branch Naming

```
feat/<story>-<description>    e.g. feat/2.1-user-registration
fix/<description>             e.g. fix/auth-token-refresh
chore/<description>           e.g. chore/update-deps
docs/<description>            e.g. docs/api-reference
```

### Commit Format

```
<type>(<scope>): <description>

feat(auth): implement user registration with bcrypt
fix(notebook): resolve optimistic lock conflict
chore(deps): update Echo to v4.12.0
```

### PR Flow

1. Branch from `main` → implement story → tests → PR → merge → auto-deploy

---

## 15. Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `PORT` | No | `8080` | Server port |
| `ENVIRONMENT` | No | `development` | `development` or `production` |
| `APP_URL` | No | `http://localhost:8080` | Public URL |
| `DATABASE_URL` | **Yes** | — | PostgreSQL connection string |
| `REDIS_URL` | No | `redis://localhost:6379` | Redis connection string |
| `JWT_SECRET` | **Yes** | — | JWT signing secret (min 32 chars) |
| `VITE_API_URL` | No | `http://localhost:8080` | Frontend API URL (build time) |
| `VITE_WS_URL` | No | `ws://localhost:8080` | Frontend WebSocket URL (build time) |

### .env.example

```bash
PORT=8080
ENVIRONMENT=development
APP_URL=http://localhost:8080
DATABASE_URL=postgresql://gobook:secret@localhost:5432/gobook?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=dev-secret-key-change-in-production-min-32-chars
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

---

## Success Metrics

### Technical KPIs

| Metric | Target |
|---|---|
| API response time (p95) | < 100ms |
| WebSocket latency | < 50ms |
| Code execution | < 5s |
| Frontend FCP | < 1.5s |
| Uptime | 99.9% |
| Test coverage (backend) | > 70% |

### User Experience KPIs

| Metric | Target |
|---|---|
| Notebook creation | < 30s |
| Sharing a notebook | < 15s |
| Joining collaboration | < 5s |
| Data loss incidents | Zero |

---

*This document is the single source of truth for building GoBook. Cursor AI should reference this for all implementation details, API contracts, data models, and architectural decisions. Stories should be implemented sequentially within each sprint.*
