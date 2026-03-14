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
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PRODUCTS_URL string = "https://api-o2o.lotuss.com/lotuss-mobile-bff/product/v4/products"

type Filter struct {
	AttributeCode string `bson:"attribute_code"`
	Label         string `bson:"label"`
	OptionLabel   string `bson:"option_label"`
	OptionValue   string `bson:"option_value"`
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

func (l *lotuss) runProducts() {
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

	sem := make(chan struct{}, 3)
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

func (l *lotuss) scrapeCategory(categoryID int) {
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

		// concurrent upsert products (sem=5)
		prodSem := make(chan struct{}, 5)
		var prodWg sync.WaitGroup
		for _, p := range products {
			sku, _ := p["sku"]
			if sku == nil {
				continue
			}
			prodWg.Add(1)
			prodSem <- struct{}{}
			go func(doc map[string]interface{}, s interface{}) {
				defer prodWg.Done()
				defer func() { <-prodSem }()
				ctx := context.Background()
				filter := bson.M{"sku": s}
				opts := options.Update().SetUpsert(true)
				if _, err := productCol.UpdateOne(ctx, filter, bson.M{"$set": doc}, opts); err != nil {
					log.Printf("failed to upsert product sku=%v: %v", s, err)
				}
			}(p, sku)
		}
		prodWg.Wait()

		// concurrent upsert filters (sem=5)
		filterSem := make(chan struct{}, 5)
		var filterWg sync.WaitGroup
		for _, f := range filters {
			filterWg.Add(1)
			filterSem <- struct{}{}
			go func(fl Filter) {
				defer filterWg.Done()
				defer func() { <-filterSem }()
				ctx := context.Background()
				filter := bson.M{"option_value": fl.OptionValue}
				opts := options.Replace().SetUpsert(true)
				if _, err := filterCol.ReplaceOne(ctx, filter, fl, opts); err != nil {
					log.Printf("failed to upsert filter %s/%s: %v", fl.AttributeCode, fl.OptionValue, err)
				}
			}(f)
		}
		filterWg.Wait()

		// random sleep 199–523ms before next page
		sleep := time.Duration(199+rand.Intn(325)) * time.Millisecond
		time.Sleep(sleep)
		page++
	}
}

func (l *lotuss) fetchProducts(categoryID, page int) ([]map[string]interface{}, []Filter, error) {
	url := fmt.Sprintf("%s?category_id=%d&page=%d&limit=15", PRODUCTS_URL, categoryID, page)
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
