# Scraper Architecture — internal/scraper/

## Overview
`internal/scraper` ดูดข้อมูลจาก external websites แล้วเก็บในรูปแบบ **raw** ลง MongoDB
โครงสร้างเป็นแบบอิสระ ไม่ strict clean architecture — แต่ละ scraper source แยกเป็นโฟลเดอร์ของตัวเอง

```
internal/scraper/
└── lotuss/              # Lotus's supermarket scraper
    ├── lotuss.go        # Bootstrap: Gin setup, MongoDB init, endpoint registration
    ├── categories.go    # ดูด category tree
    ├── products.go      # ดูด products รายหมวด (paginated)
    └── images.go        # Download product/category images
```

## Database Schema (MongoDB: `lotuss_raw`)

| Collection | Unique Key | Description |
|------------|-----------|-------------|
| `categories` | `id` (int) | Category tree (flattened) จาก Lotus API |
| `products` | `sku` (string) | Raw product maps จาก Lotus API |
| `filters` | `option_value` | Attribute filters (สี, ขนาด, ฯลฯ) |

**Storage:** Images ดาวน์โหลดลง `storage/` โดย mirror path จาก URL

---

## Lotuss Struct (lotuss.go)
```go
type Lotuss struct {
    db                 *mongo.Database   // lotuss_raw database
    skipExistingImages bool
}
```
- `NewLotuss(cfg)` → connect MongoDB, init struct
- `RunAPI(port)` → เปิด Gin server พร้อม endpoint `GET /scrap/all`

**`/scrap/all` handler** — เรียก 4 goroutines พร้อมกัน:
```go
go l.RunCategories()
go l.RunProducts()
go l.RunImages()
go l.RunCategoryImages()
```

---

## Concurrency Pattern (Semaphore + WaitGroup)

ใช้ทั่วทั้ง scraper เพื่อจำกัดจำนวน concurrent goroutines:

```go
const maxConcurrency = 10
sem := make(chan struct{}, maxConcurrency)
var wg sync.WaitGroup

for _, item := range items {
    wg.Add(1)
    sem <- struct{}{}            // acquire slot
    go func(i Item) {
        defer wg.Done()
        defer func() { <-sem }() // release slot
        // do work
    }(item)
}
wg.Wait()
```

---

## Upsert Strategy (MongoDB)

ทุก collection ใช้ `UpdateOne` + `upsert: true`:

```go
filter := bson.M{"id": category.ID}
update := bson.M{
    "$set": bson.M{
        "name":       category.Name,
        "updated_at": time.Now(),
        // ... all fields
    },
    "$setOnInsert": bson.M{
        "created_at": time.Now(),
    },
}
opts := options.Update().SetUpsert(true)
col.UpdateOne(ctx, filter, update, opts)
```

- `$set` — อัปเดตทุกครั้ง
- `$setOnInsert` — set เฉพาะตอน insert ครั้งแรก (`created_at`)

---

## Bulk Write Batching Pattern

Products ใช้ `BulkWrite` เพื่อประสิทธิภาพ แบ่งเป็น batch:

```go
func bulkWrite(ctx, col, models []mongo.WriteModel, batchSize int) error {
    for i := 0; i < len(models); i += batchSize {
        end := min(i+batchSize, len(models))
        _, err := col.BulkWrite(ctx, models[i:end],
            options.BulkWrite().SetOrdered(false))
        if err != nil { return err }
    }
    return nil
}
```

- `SetOrdered(false)` — parallel execution, ไม่หยุดเมื่อ error
- Batch size: 50 documents

---

## Rate Limiting (products.go)

ใส่ random delay ระหว่าง page requests เพื่อหลีกเลี่ยง rate limit:

```go
time.Sleep(time.Duration(97+rand.Intn(52)) * time.Millisecond) // 97–149ms
```

---

## Image Download Pattern (images.go)

```
URL: https://cdn.lotuss.com/media/catalog/product/a/b/image.jpg
 ↓ parse URL path
Local: storage/media/catalog/product/a/b/image.jpg
```

Steps:
1. Query MongoDB → รวบ URL ทั้งหมด
2. Semaphore goroutines (10 concurrent)
3. Check `SKIP_EXISTING_IMAGES` → skip ถ้าไฟล์มีอยู่แล้ว
4. `http.Get(url)` → `io.Copy(file, resp.Body)`

---

## Category Scraping Flow (categories.go)

1. `GET https://api-o2o.lotuss.com/lotuss-mobile-bff/product/v4/categories`
2. Flatten tree recursively (Children → flat list พร้อม `level` และ `path`)
3. Upsert ทุก category พร้อม semaphore (10 concurrent)

**Category fields:** `id`, `name`, `image`, `slug`, `level`, `path`, `is_hle`, `is_b2b`, `is_on_demand`, `is_in_menu`, `is_online`, `updated_at`, `created_at`

---

## Product Scraping Flow (products.go)

1. Query MongoDB: level-1 categories
2. For each category (10 concurrent):
   - Loop pages (1, 2, 3...) จนหมด
   - `GET .../products?category_id=X&page=N&size=50`
   - Upsert products (bulk, batch 50)
   - Extract filters → upsert `filters` collection
   - Random sleep ระหว่าง pages

---

## How to Add a New Scraper Source (e.g., "bigc")

1. สร้างโฟลเดอร์ `internal/scraper/bigc/`
2. สร้าง `bigc.go` — struct + `NewBigC(cfg)` + `RunAPI(port)` + endpoint
3. สร้างไฟล์ scraping logic ตามต้องการ (categories, products, images)
4. เพิ่ม config fields ใน `internal/shared/config.go` ถ้าต้องการ
5. สร้าง entry point `cmd/scraper/bigc/main.go` หรือเพิ่มใน existing scraper cmd
6. Database: ใช้ database ใหม่เช่น `bigc_raw` เพื่อแยก raw data
