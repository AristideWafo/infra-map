# InfraMaps — Agent Context

> Read this file first. Every decision made here is final for this codebase.

## What this project is

InfraMaps is a **topology-first infrastructure visualization tool**. It connects to existing observability infrastructure (Prometheus, K8s API, Docker socket, Loki) and renders a spatial, navigable map of the infrastructure — from datacenter down to individual containers.

**Single binary Go backend. React frontend. Zero new agents at client sites.**

## Architecture in one paragraph

The Go backend scrapes multiple optional sources every 30 seconds, normalizes everything into a `UnifiedNode` tree, computes stable layout positions, and serves via REST + WebSocket. The frontend polls the tree and renders a 2D grid (default) or 3D spatial map (Phase 4). No database — Prometheus holds history. No new agents — existing Prometheus exporters cover VMs.

## Hard rules — Never violate

```
MUST  Prometheus is the only mandatory source. K8s, Docker, Loki are optional.
MUST  No database. In-memory cache with 30s TTL only.
MUST  Layout is deterministic: same input → same positions, always.
MUST  Scraper interface: every source implements Name() / Scrape() / Health().
MUST  Every scraper must be testable without a real connection (mockable interface).
MUST  context.Context in every external call. Timeout enforced.
MUST  Errors returned, never panic() in production code.
MUST  Memory values in MB everywhere (never bytes in API responses).

NEVER Add a database (SQLite, Redis, BoltDB — any of them).
NEVER Add write actions to production (restart pod, scale, exec shell) — out of scope.
NEVER Generate non-deterministic layout positions (no random, no physics simulation).
NEVER Put business logic in HTTP handlers — handlers call services only.
NEVER Fetch data in React components — always in hooks.
```

## Tech stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Backend language | Go | 1.22+ |
| HTTP framework | Gin | latest |
| K8s client | client-go | v0.29+ |
| Prometheus client | prometheus/client_golang | latest |
| WebSocket | gorilla/websocket | latest |
| Frontend | React + TypeScript | 19 / 5.x |
| State | Zustand | 4.x |
| Charts | Recharts | 2.x |
| Build tool | Vite | 5.x |
| Tests (Go) | testing + testify | stdlib + v1.x |
| Tests (React) | Vitest + RTL | latest |
| Packaging | Docker (multi-stage) + Helm | — |

## Project structure

```
infra-maps/
├── CLAUDE.md                    ← You are here
├── ROADMAP.md
├── docs/
│   ├── ARCHITECTURE.md
│   ├── API_CONTRACT.md
│   ├── DATA_MODELS.md
│   ├── DESIGN_SYSTEM.md
│   ├── BACKEND_GUIDE.md
│   ├── FRONTEND_GUIDE.md
│   ├── DEPLOYMENT.md
│   └── adr/
│       ├── ADR-001-multi-source.md
│       ├── ADR-002-no-database.md
│       └── ADR-003-layout-algorithm.md
├── infra-maps-api/              ← Go backend
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── api/
│   │   ├── scraper/
│   │   ├── layout/
│   │   ├── resolver/
│   │   ├── cache/
│   │   └── models/
│   ├── Dockerfile
│   └── go.mod
├── infra-maps-ui/               ← React frontend
│   ├── src/
│   │   ├── api/
│   │   ├── store/
│   │   ├── hooks/
│   │   ├── views/
│   │   └── components/
│   ├── package.json
│   └── vite.config.ts
└── helm/infra-maps/             ← Helm chart
```

## Key docs to read before coding

- `docs/DATA_MODELS.md` → canonical types (Go + TS). Read before touching any struct.
- `docs/API_CONTRACT.md` → endpoint contracts. Do not add undocumented endpoints.
- `docs/ARCHITECTURE.md` → system understanding + multi-source capability matrix.
- `docs/DESIGN_SYSTEM.md` → colors, sizes, component specs. Use CSS variables only.
- `docs/BACKEND_GUIDE.md` → Go patterns, scraper convention, testing.
- `docs/FRONTEND_GUIDE.md` → React patterns, Zustand store, hook conventions.

## Current phase

See `ROADMAP.md` for the active phase and exit conditions.
