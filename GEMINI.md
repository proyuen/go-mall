# Project: go-mall (Monolithic RESTful API)

## Tech Stack
*   **Frontend**: TypeScript, React, Vite, Ant Design (or Tailwind CSS + Shadcn/ui), Zustand (State Management).
*   **Backend**: Go (Golang), Gin (HTTP Framework), GORM (ORM).
*   **Database**: PostgreSQL.
*   **Cache & Lock**: Redis (Cache, Distributed Lock, Lua Scripts).
*   **Message Queue**: RabbitMQ (Async Orders, Traffic Shaping).
*   **Search**: Elasticsearch (Optional, later addition).
*   **DevOps**: Docker, Docker Compose.
*   **Testing**: Go testing, Testify, Gomock, K6 (Load Testing).

## Project Core Features (MVP)
1.  **User System**: JWT Dual Token Auth, RBAC.
2.  **Product System**: SPU/SKU Design, High-performance list & detail APIs.
3.  **High Concurrency Flash Sale (Core Highlight)**:
    *   Redis + Lua for inventory pre-deduction.
    *   RabbitMQ for async order creation (traffic shaping).
    *   Redis Distributed Lock to prevent overselling/duplicate requests.
4.  **Architecture**:
    *   **Monolithic RESTful API**: Single binary, clean layered architecture (`Handler` -> `Service` -> `Repository` -> `DB`).
    *   **Gin**: High-performance HTTP web framework.
5.  **Engineering**: Complete Unit Test coverage, Swagger API Docs.

## Workflow
**Phase 1: Architecture Design & Environment**
1.  Design Project Directory (Monolithic Standard Layout).
2.  `docker-compose.yml` for PG, Redis, RabbitMQ.
3.  Database Schema (User, SKU, SPU, Order).

**Phase 2: Core Business Logic**
1.  **User Module**: Registration, Login (JWT).
2.  **Product Module**: CRUD, List, Detail.
3.  **Order Module**: Basic Order Creation.

**Phase 3: High Concurrency Logic**
1.  **Flash Sale**: Redis Lua Script, RabbitMQ Producer/Consumer.
2.  **Optimization**: Caching strategies.

**Phase 4: Frontend**
1.  Vite + React + TS Setup.
2.  UI Implementation.

**Phase 5: Test & Optimize**
1.  Benchmarks.
2.  Unit Tests.