package lotuss

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/siwarats/price-tag/internal/shared"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DB_NAME             string = "lotuss_raw"
	CATEGORY_COLLECTION string = "categories"
	PRODUCT_COLLECTION  string = "products"
	FILTER_COLLECTION   string = "filters"
)

type Lotuss struct {
	db                 *mongo.Database
	skipExistingImages bool
}

func NewLotuss(cfg *shared.Config) *Lotuss {
	ctx := context.Background()
	clientOpts := options.Client().ApplyURI(cfg.MONGODB_URI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping MongoDB: %v", err)
	}
	return &Lotuss{
		db:                 client.Database(DB_NAME),
		skipExistingImages: cfg.SKIP_EXISTING_IMAGES,
	}
}

func (l *Lotuss) RunAPI(port string) {
	r := gin.Default()
	r.GET("/scrap/all", func(c *gin.Context) {
		go func() {
			l.RunCategories()
			l.RunProducts()
			l.RunImages()
			l.RunCategoryImages()
		}()
		c.Status(http.StatusOK)
	})
	// r.GET("/scrap/categories", func(c *gin.Context) { go l.RunCategories(); c.Status(http.StatusOK) })
	// r.GET("/scrap/products", func(c *gin.Context) { go l.RunProducts(); c.Status(http.StatusOK) })
	// r.GET("/scrap/images", func(c *gin.Context) { go l.RunImages(); c.Status(http.StatusOK) })
	// r.GET("/scrap/category-images", func(c *gin.Context) { go l.RunCategoryImages(); c.Status(http.StatusOK) })
	r.Run(":" + port)
}
