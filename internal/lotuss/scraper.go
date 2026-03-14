package lotuss

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/siwarats/price-tag/pkg"
)

const (
	DB_NAME             string = "lotuss_raw"
	CATEGORY_COLLECTION string = "categories"
	PRODUCT_COLLECTION  string = "products"
	FILTER_COLLECTION   string = "filters"
)

type lotuss struct {
	db                 *mongo.Database
	skipExistingImages bool
}

func NewLotuss(cfg *pkg.Config) *lotuss {
	ctx := context.Background()
	clientOpts := options.Client().ApplyURI(cfg.MONGODB_URI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping MongoDB: %v", err)
	}
	return &lotuss{
		db:                 client.Database(DB_NAME),
		skipExistingImages: cfg.SKIP_EXISTING_IMAGES,
	}
}

func (l *lotuss) Run() {
	// categories, err := l.fetchCategories()
	// if err != nil {
	// 	log.Fatalf("failed to fetch categories: %v", err)
	// }

	// flat := flattenCategories(categories)
	// log.Printf("total categories: %d", len(flat))

	// col := l.db.Collection(CATEGORY_COLLECTION)
	// l.upsertCategoriesConcurrent(col, flat, 10)
	// log.Println("done inserting categories")

	l.runProducts()
	// l.runImages()
	// l.runCategoryImages()
}
