# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common commands

### Full stack with Docker Compose
- `docker compose up -d --build`
- `docker compose ps`
- `docker compose logs -f backend`
- `docker compose logs -f frontend`
- `docker compose down`
- `docker compose down -v`

### Infra-only local development
- `docker compose up -d db redis mq`
- `docker compose exec db mysql -uroot -proot123 smart_campus`

### Backend (`backend/` is the Go module root)
- `cd backend && go run ./cmd/server`
- `cd backend && go build ./cmd/server`
- `cd backend && go test ./...`
- Single Go test (when tests exist): `cd backend && go test ./path/to/package -run TestName`

### Frontend
- `cd frontend && npm install`
- `cd frontend && npm run dev`
- `cd frontend && npm run build`
- `cd frontend && npm run lint`
- `cd frontend && npm run preview`

### Local URLs
- Frontend via Compose: `http://localhost`
- Frontend via Vite: `http://localhost:3000`
- Backend health check: `http://localhost:8080/health`
- RabbitMQ UI: `http://localhost:15672` (`guest` / `guest`)

## Current implementation snapshot

- `ARCHITECTURE.md` and `DEV_DOC.md` describe a larger target platform with gateway/microservices and MQ-heavy flows. The checked-in code today is a single Go API server plus a React SPA. Prefer the code over those docs when they disagree.
- Backend startup auto-runs GORM `AutoMigrate`, so a fresh database gets schema automatically.
- The repo has no built-in seed script or admin UI. After a fresh boot, empty spaces/activities usually mean the database has no demo rows yet rather than a frontend bug.

## Backend architecture

- Entry point: `backend/cmd/server/main.go`.
- The API is one Gin application with route groups under `/api/v1`.
- Main handler modules are:
  - `backend/internal/handler/user`
  - `backend/internal/handler/space`
  - `backend/internal/handler/seckill`
  - `backend/internal/handler/order`
- Business logic lives mostly inside handlers rather than separate service/repository layers. When adding behavior, expect to update handler code directly and then wire routes in `main.go`.
- Auth uses JWT via `backend/internal/pkg/jwt` and `backend/internal/pkg/middleware`; protected handlers read `userID` and `role` from Gin context.
- `middleware.ErrorMiddleware()` exists but is not currently registered in `main.go`; most handlers return JSON errors inline.

## Domain model and critical flows

- Core models live in `backend/internal/model`.
- Space resources use a polymorphic schema:
  - `resources` is the shared base table.
  - `academic_spaces` stores academic-space-specific settings.
  - `sports_facilities` stores sports-specific settings.
- Booking availability and conflict state live in `time_slots`.
- Space booking (`backend/internal/handler/space/space.go`) works by:
  1. loading the resource and slot,
  2. checking slot status and user credit score,
  3. claiming the slot with an optimistic-lock `version` update,
  4. creating an `orders` row plus `order_items` detail row.
- Orders are shared by both booking and seckill flows. `order_type` distinguishes `space` vs `activity`, while `order_items` stores the resource/activity-specific detail.
- Order payment/cancellation (`backend/internal/handler/order/order.go`) uses a DB transaction; payment adds a row lock plus order `version` check before setting status.
- Activity seckill (`backend/internal/handler/seckill/seckill.go`) is the only place Redis is used in current code:
  - an inline Lua script does atomic stock decrement and one-user-one-purchase checks,
  - then the handler creates the order in MySQL,
  - then it decrements `activities.remaining_tickets` in MySQL.
- RabbitMQ is configured in Compose and config loading, but the current Go code does not publish or consume MQ messages yet.
- The order status enum includes `pending`, `confirmed`, `paid`, `cancelled`, `no_show`, and `completed`, but the current handlers mainly transition through `pending`, `paid`, and `cancelled`.

## Frontend architecture

- App bootstrap: `frontend/src/main.tsx`.
- Routing: `frontend/src/App.tsx`.
  - `/login` and `/register` are public.
  - Everything under `/` is wrapped in `ProtectedRoute` and requires auth.
- Layout and top-level navigation live in `frontend/src/components/layout/Layout.tsx`.
- Persisted auth state lives in `frontend/src/stores/authStore.ts` using Zustand `persist`; token and user are stored in local storage.
- All HTTP calls are centralized in `frontend/src/services/api.ts`:
  - Axios base URL is `/api/v1`.
  - A request interceptor injects `Authorization: Bearer <token>`.
  - A `401` response logs the user out and redirects to `/login`.
- Main screens are page-driven rather than using a heavier client-state architecture:
  - `frontend/src/pages/Spaces.tsx` loads spaces and slots, then books through `spaceApi`.
  - `frontend/src/pages/Activities.tsx` lists activities and triggers seckill directly.
  - `frontend/src/pages/Orders.tsx` lists orders and calls pay/cancel actions.
- Styling is mostly Tailwind utility classes in page/components rather than a large shared component system.

## Runtime and proxy behavior

- Vite dev server proxies `/api` to `http://localhost:8080` (`frontend/vite.config.ts`).
- The production frontend container serves the SPA from Nginx and proxies `/api/` to `backend:8080` (`frontend/nginx.conf`).
- The frontend assumes the backend API prefix remains `/api/v1`.

## Practical repo-specific notes

- For first-time manual testing, start at `http://localhost/register` because the home route is auth-protected.
- If the UI renders but lists are empty, check MySQL content before debugging React state; `README.md` contains demo SQL for `resources`, `time_slots`, and `activities`.
- There are currently no committed Go `*_test.go` files and no frontend test script in `frontend/package.json`, so validation is mainly backend build/test plus frontend lint/build.
