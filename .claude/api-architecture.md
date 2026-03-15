# API Architecture — internal/api/

## Overview
`internal/api` ใช้ **Clean Architecture** แบบ strict 3 ชั้น ทุก feature แยกเป็นโฟลเดอร์ย่อยของตัวเอง

```
internal/api/
├── api.go              # Bootstrap: สร้าง Gin engine, wire DI, register routes
├── entity/             # Domain models (shared across layers)
├── di/                 # Dependency Injection container
│   ├── repository.go
│   ├── usecase.go
│   └── route.go
├── route/              # Layer 1: HTTP handlers (controllers)
│   └── product/
├── usecase/            # Layer 2: Business logic
│   └── product/
└── repository/         # Layer 3: Data access (MongoDB)
    └── product/
```

## 3-Layer Architecture

```
HTTP Request
     ↓
[ Route / Handler ]   ← รับ *gin.Context, parse input, return JSON
     ↓
[ UseCase ]           ← Business logic, orchestration, validation
     ↓
[ Repository ]        ← MongoDB queries, data mapping
     ↓
  MongoDB
```

### Layer 1: Route (`route/<feature>/`)
- **ไฟล์:** `<feature>.go` (interface + constructor), `get_<entity>.go`, `get_<entities>.go`
- **Interface pattern:**
  ```go
  type ProductRoute interface {
      RegisterRoutes()
      GetProduct(ctx *gin.Context)
      GetProducts(ctx *gin.Context)
  }
  ```
- **Constructor:** `NewProductRoute(rg *gin.RouterGroup, uc usecase.ProductUseCase) ProductRoute`
- **Route registration:** สร้าง sub-group แล้ว map method → handler
  ```go
  func (r *productRoute) RegisterRoutes() {
      g := r.routerGroup.Group("/products")
      g.GET("/:id", r.GetProduct)
      g.GET("/", r.GetProducts)
  }
  ```
- Handler parse input → เรียก usecase → return `ctx.JSON(...)`

### Layer 2: UseCase (`usecase/<feature>/`)
- **ไฟล์:** `<feature>.go` (interface + constructor), method ละไฟล์
- **Interface pattern:**
  ```go
  type ProductUseCase interface {
      GetProduct(id string) (*entity.Product, error)
      GetProducts() ([]entity.Product, error)
  }
  ```
- **Constructor:** `NewProductUseCase(repo repository.ProductRepository) ProductUseCase`
- ชั้นนี้ไม่รู้จัก `*gin.Context` เลย — รับ/return domain types เท่านั้น

### Layer 3: Repository (`repository/<feature>/`)
- **ไฟล์:** `<feature>.go` (interface + constructor), method ละไฟล์
- **Interface pattern:**
  ```go
  type ProductRepository interface {
      GetProduct(id string) (*entity.Product, error)
      GetProducts() ([]entity.Product, error)
  }
  ```
- **Constructor:** `NewProductRepository() ProductRepository`
- รับ MongoDB client/collection เป็น dependency (ผ่าน constructor)

### Entity (`entity/`)
- Plain Go structs — ไม่มี logic
- ใช้ร่วมกันทุกชั้น (route/usecase/repository ล้วน import entity)
- ตัวอย่าง: `entity.Product`

---

## Dependency Injection (di/)

Manual DI — wire จาก bottom ขึ้น top:

```
di.NewRepository()                        → *repository (holds ProductRepository)
di.NewUseCase(diRepository)               → *useCase (holds ProductUseCase)
di.NewRoute(routerGroup, diUseCase)       → *route (holds ProductRoute)
```

**api.go wiring:**
```go
diRepository := di.NewRepository()
diUseCase    := di.NewUseCase(diRepository)
diRoute      := di.NewRoute(apiV1, diUseCase)
diRoute.GetProductRoute().RegisterRoutes()
```

เมื่อเพิ่ม feature ใหม่:
1. เพิ่ม field ใน `di/repository.go`, `di/usecase.go`, `di/route.go`
2. เรียก `.RegisterRoutes()` ใน `api.go`

---

## How to Add a New Feature (e.g., "category")

1. **Entity:** สร้าง `entity/category.go` → define `Category` struct
2. **Repository:** สร้าง `repository/category/category.go` + method files
   - Define `CategoryRepository` interface
   - Implement `categoryRepository` struct + constructor
3. **UseCase:** สร้าง `usecase/category/category.go` + method files
   - Define `CategoryUseCase` interface
   - Implement `categoryUseCase` struct รับ `CategoryRepository`
4. **Route:** สร้าง `route/category/category.go` + method files
   - Define `CategoryRoute` interface
   - Implement `categoryRoute` struct รับ `CategoryUseCase`
   - Register routes ใน `RegisterRoutes()`
5. **DI:** เพิ่ม field + getter ใน `di/repository.go`, `di/usecase.go`, `di/route.go`
6. **api.go:** เรียก `diRoute.GetCategoryRoute().RegisterRoutes()`

---

## API Routes (current)
```
GET /api/v1/products/:id   → ProductRoute.GetProduct
GET /api/v1/products/      → ProductRoute.GetProducts
```

## Gin Setup (api.go)
```go
engine := gin.Default()
apiV1  := engine.Group("/api/v1")
// → DI wiring here
engine.Run(":" + port)
```
