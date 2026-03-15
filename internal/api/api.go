package api

import (
	"github.com/gin-gonic/gin"
	"github.com/siwarats/price-tag/internal/api/di"
	"github.com/siwarats/price-tag/internal/shared"
)

type API struct {
	cfg *shared.Config
}

func NewAPI(
	cfg *shared.Config,
) *API {
	return &API{
		cfg: cfg,
	}
}

func (a *API) ServeHttp(port string) {
	r := gin.Default()
	apiV1 := r.Group("/api/v1")

	diRepository := di.NewRepository()
	diUseCase := di.NewUseCase(diRepository)
	diRoute := di.NewRoute(apiV1, diUseCase)

	diRoute.GetProductRoute().RegisterRoutes()

	r.Run(":" + port)
}
