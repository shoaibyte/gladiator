# Gladiator

Interactive Go notebooks in the browser — inspired by Elixir LiveBook. The Go gopher is your gladiator in the arena of code.

## Stack

- **Backend:** Go, Echo, Ent (PostgreSQL), Redis, yaegi (Go interpreter), JWT auth, WebSocket
- **Frontend:** React, TypeScript, Vite, Tailwind, Monaco Editor, Zustand, React Query

## Quick start

```bash
cp .env.example .env
make dev
```

- API: http://localhost:8080  
- Frontend: http://localhost:5173  
- Health: http://localhost:8080/health  

## Build

```bash
make install-deps
make generate   # Ent codegen
make build      # frontend + backend binary in bin/server
```

## Deploy

- Docker: `make docker-build` then run the image with `DATABASE_URL`, `REDIS_URL`, `JWT_SECRET`.
- Render: use `render.yaml` (DB + Redis + web service).

See [Gladiator-Development-Documentation.md](Gladiator-Development-Documentation.md) for full spec and API.
