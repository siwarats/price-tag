package route

import (
	"github.com/gin-gonic/gin"
	usecase "github.com/siwarats/price-tag/internal/api/usecase/product"
)

type ProductRoute interface {
	RegisterRoutes()
	GetProduct(ctx *gin.Context)
	GetProducts(ctx *gin.Context)
}

type productRoute struct {
	routerGroup *gin.RouterGroup
	usecase     usecase.ProductUseCase
}

func NewProductRoute(
	routerGroup *gin.RouterGroup,
	productUseCase usecase.ProductUseCase,
) *productRoute {

	return &productRoute{
		routerGroup: routerGroup.Group("/products"),
		usecase:     productUseCase,
	}
}

func (h *productRoute) RegisterRoutes() {
	h.routerGroup.GET("/:id", h.GetProduct)
	h.routerGroup.GET("/", h.GetProducts)
}
