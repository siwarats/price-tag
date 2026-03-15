package lotuss

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PRODUCTS_URL string = "https://api-o2o.lotuss.com/lotuss-mobile-bff/product/v4/products"

type Filter struct {
	AttributeCode string    `bson:"attribute_code"`
	Label         string    `bson:"label"`
	OptionLabel   string    `bson:"option_label"`
	OptionValue   string    `bson:"option_value"`
	UpdatedAt     time.Time `bson:"updated_at"`
}

type filterRaw struct {
	AttributeCode string `json:"attributeCode"`
	Label         string `json:"label"`
	Options       []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"options"`
}

type productsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Products []map[string]interface{} `json:"products"`
		Filters  []filterRaw              `json:"filters"`
	} `json:"data"`
}

func (l *Lotuss) RunProducts() {
	ctx := context.Background()
	cur, err := l.db.Collection(CATEGORY_COLLECTION).Find(ctx, bson.M{"level": 1})
	if err != nil {
		log.Fatalf("failed to query level-1 categories: %v", err)
	}
	defer cur.Close(ctx)

	var level1 []Category
	if err := cur.All(ctx, &level1); err != nil {
		log.Fatalf("failed to decode level-1 categories: %v", err)
	}
	log.Printf("level-1 categories: %d", len(level1))

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for _, cat := range level1 {
		wg.Add(1)
		sem <- struct{}{}
		go func(c Category) {
			defer wg.Done()
			defer func() { <-sem }()
			l.scrapeCategory(c.ID)
		}(cat)
	}

	wg.Wait()
	log.Println("done scraping products")
}

func (l *Lotuss) scrapeCategory(categoryID int) {
	productCol := l.db.Collection(PRODUCT_COLLECTION)
	filterCol := l.db.Collection(FILTER_COLLECTION)

	page := 1
	for {
		products, filters, err := l.fetchProducts(categoryID, page)
		if err != nil {
			log.Printf("[cat=%d page=%d] fetch error: %v", categoryID, page, err)
			break
		}
		if len(products) == 0 {
			log.Printf("[cat=%d] no products on page %d, done", categoryID, page)
			break
		}

		log.Printf("[cat=%d page=%d] got %d products, %d filters", categoryID, page, len(products), len(filters))

		// bulk upsert products in batches of 50
		var productModels []mongo.WriteModel
		for _, p := range products {
			sku, _ := p["sku"]
			if sku == nil {
				continue
			}
			p["category_id"] = categoryID
			p["updated_at"] = time.Now()
			productModels = append(productModels, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"sku": sku}).
				SetUpdate(bson.M{
					"$set":         p,
					"$setOnInsert": bson.M{"created_at": time.Now()},
				}).
				SetUpsert(true))
		}
		if err := bulkWrite(context.Background(), productCol, productModels, 50); err != nil {
			log.Printf("[cat=%d page=%d] bulk upsert products error: %v", categoryID, page, err)
		}

		// bulk upsert filters in batches of 50
		var filterModels []mongo.WriteModel
		for _, f := range filters {
			fl := f
			fl.UpdatedAt = time.Now()
			filterModels = append(filterModels, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"option_value": fl.OptionValue}).
				SetUpdate(bson.M{
					"$set":         fl,
					"$setOnInsert": bson.M{"created_at": time.Now()},
				}).
				SetUpsert(true))
		}
		if err := bulkWrite(context.Background(), filterCol, filterModels, 50); err != nil {
			log.Printf("[cat=%d page=%d] bulk upsert filters error: %v", categoryID, page, err)
		}

		// random sleep 97–149ms before next page
		sleep := time.Duration(97+rand.Intn(52)) * time.Millisecond
		time.Sleep(sleep)
		page++
	}
}

func bulkWrite(ctx context.Context, col *mongo.Collection, models []mongo.WriteModel, batchSize int) error {
	opts := options.BulkWrite().SetOrdered(false)
	for i := 0; i < len(models); i += batchSize {
		end := i + batchSize
		if end > len(models) {
			end = len(models)
		}
		if _, err := col.BulkWrite(ctx, models[i:end], opts); err != nil {
			return err
		}
	}
	return nil
}

func (l *Lotuss) fetchProducts(categoryID, page int) ([]map[string]interface{}, []Filter, error) {
	url := fmt.Sprintf("%s?category_id=%d&page=%d&limit=50", PRODUCTS_URL, categoryID, page)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var result productsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("decode response: %w", err)
	}
	if result.Code != 200 {
		return nil, nil, fmt.Errorf("unexpected response code: %d message: %s", result.Code, result.Message)
	}

	var filters []Filter
	for _, fr := range result.Data.Filters {
		for _, opt := range fr.Options {
			filters = append(filters, Filter{
				AttributeCode: fr.AttributeCode,
				Label:         fr.Label,
				OptionLabel:   opt.Label,
				OptionValue:   opt.Value,
			})
		}
	}

	return result.Data.Products, filters, nil
}
