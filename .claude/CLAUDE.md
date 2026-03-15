# CLAUDE.md — price-tag Project Overview

## Project Purpose
ระบบดูดข้อมูลราคาสินค้าจาก external websites (เริ่มจาก Lotus's) แล้วเปิดเป็น REST API ให้เว็บใช้งาน

## Tech Stack
- **Language:** Go 1.23
- **HTTP Framework:** `gin-gonic/gin` v1.10.0
- **Database:** MongoDB (`mongo-driver` v1.17.3)
- **Config:** `godotenv` — โหลด `.env` อัตโนมัติ
- **Module:** `github.com/siwarats/price-tag`

## Repository Structure
```
price-tag/
├── cmd/
│   ├── api/main.go          # Entry point: REST API server
│   └── scraper/main.go      # Entry point: Scraper service
├── internal/
│   ├── api/                 # REST API (Clean Architecture) → see api-architecture.md
│   ├── scraper/             # Web scrapers → see scraper-architecture.md
│   └── shared/              # Shared utilities (config)
├── storage/                 # Downloaded images (gitignored)
├── .env                     # Runtime secrets (gitignored)
└── .env.example             # Template for environment variables
```

## Two Independent Services
| Service | Entry Point | Port Env | Role |
|---------|-------------|----------|------|
| API | `cmd/api/main.go` | `API_PORT` | Serve REST endpoints to web clients |
| Scraper | `cmd/scraper/main.go` | `SCRAPER_LOTUSS_PORT` | Fetch & sync raw data from external sites |

ทั้งสองคุยกันผ่าน **MongoDB** เท่านั้น (ไม่มี internal RPC/messaging)

## Environment Variables (`.env`)
```
MONGODB_URI=             # MongoDB connection string
API_PORT=                # Port for REST API
SCRAPER_LOTUSS_PORT=     # Port for Lotus scraper API
LOTUSS_SCRAPER_URL=      # Base URL for Lotus's public API
SKIP_EXISTING_IMAGES=    # "true" = skip re-downloading existing images
```
Config โหลดผ่าน `internal/shared/config.go` → `shared.NewConfig()`

## internal/shared/
โฟลเดอร์นี้เก็บของที่ใช้ร่วมกันระหว่าง `internal/api` และ `internal/scraper`

| File | Purpose |
|------|---------|
| `config.go` | `Config` struct + `NewConfig()` — อ่าน env แล้ว return struct |

> เมื่อต้องการเพิ่ม shared utility ให้ใส่ที่นี่ เช่น shared error types, helpers

## Sub-Documents
- [api-architecture.md](api-architecture.md) — Clean Architecture pattern ของ `internal/api`
- [scraper-architecture.md](scraper-architecture.md) — Scraper patterns ของ `internal/scraper`

## Key Conventions
- ทุก service bootstrap ผ่าน `NewX(cfg)` constructor รับ `*shared.Config`
- ไม่มี DI framework — wire dependencies ด้วยมือ (bottom-up)
- Interface ทุกชั้น — Route, UseCase, Repository ต้องมี interface ก่อนเสมอ
- Error handling: `fmt.Errorf("context: %w", err)` + log ที่ call site
- MongoDB collection ของ scraper อยู่ใน database `lotuss_raw`
