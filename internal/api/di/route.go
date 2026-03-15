package di

import (
	"github.com/gin-gonic/gin"
	productRoute "github.com/siwarats/price-tag/internal/api/route/product"
)

type Route interface {
	GetProductRoute() productRoute.ProductRoute
}

type route struct {
	product productRoute.ProductRoute
}

func NewRoute(
	rootRouterGroup *gin.RouterGroup,
	usecase *useCase,
) *route {
	return &route{
		product: productRoute.NewProductRoute(
			rootRouterGroup,
			usecase.product,
		),
	}
}

// Implementation of Route interface

func (r *route) GetProductRoute() productRoute.ProductRoute {
	return r.product
}
