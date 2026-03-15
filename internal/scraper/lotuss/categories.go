package lotuss

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CATEGORIES_URL string = "https://api-o2o.lotuss.com/lotuss-mobile-bff/product/v4/categories?seller_id=2"

type Category struct {
	ID         int        `json:"id" bson:"id"`
	Name       string     `json:"name" bson:"name"`
	Image      string     `json:"image" bson:"image"`
	Slug       string     `json:"slug" bson:"slug"`
	IsHle      int        `json:"is_hle" bson:"is_hle"`
	IsB2b      int        `json:"is_b2b" bson:"is_b2b"`
	IsOnDemand int        `json:"is_on_demand" bson:"is_on_demand"`
	IsInMenu   int        `json:"is_in_menu" bson:"is_in_menu"`
	IsOnline   int        `json:"is_online" bson:"is_online"`
	Level     int        `json:"level" bson:"level"`
	Path      string     `json:"path" bson:"path"`
	Children  []Category `json:"children" bson:"-"`
	UpdatedAt time.Time  `json:"-" bson:"updated_at"`
}

type categoriesResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Children []Category `json:"children"`
	} `json:"data"`
}

func (l *Lotuss) fetchCategories() ([]Category, error) {
	req, err := http.NewRequest(http.MethodGet, CATEGORIES_URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var result categoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if result.Code != 200 {
		return nil, fmt.Errorf("unexpected response code: %d message: %s", result.Code, result.Message)
	}

	return result.Data.Children, nil
}

func flattenCategories(categories []Category) []Category {
	var flat []Category
	for _, cat := range categories {
		children := cat.Children
		cat.Children = nil
		flat = append(flat, cat)
		flat = append(flat, flattenCategories(children)...)
	}
	return flat
}

func (l *Lotuss) RunCategories() {
	categories, err := l.fetchCategories()
	if err != nil {
		log.Fatalf("failed to fetch categories: %v", err)
	}

	flat := flattenCategories(categories)
	log.Printf("total categories: %d", len(flat))

	col := l.db.Collection(CATEGORY_COLLECTION)
	l.upsertCategoriesConcurrent(col, flat, 10)
	log.Println("done inserting categories")
}

func (l *Lotuss) upsertCategoriesConcurrent(col *mongo.Collection, categories []Category, concurrency int) {
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, cat := range categories {
		wg.Add(1)
		sem <- struct{}{}
		go func(c Category) {
			defer wg.Done()
			defer func() { <-sem }()

			c.UpdatedAt = time.Now()
			ctx := context.Background()
			filter := bson.M{"id": c.ID}
			update := bson.M{
				"$set":         c,
				"$setOnInsert": bson.M{"created_at": time.Now()},
			}
			opts := options.Update().SetUpsert(true)
			if _, err := col.UpdateOne(ctx, filter, update, opts); err != nil {
				log.Printf("failed to upsert category id=%d: %v", c.ID, err)
			} else {
				log.Printf("upserted category id=%d name=%s level=%d", c.ID, c.Name, c.Level)
			}
		}(cat)
	}

	wg.Wait()
}
