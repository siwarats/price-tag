package main

import (
	"github.com/siwarats/price-tag/internal/api"
	"github.com/siwarats/price-tag/internal/shared"
)

func main() {
	cfg := shared.NewConfig()

	api := api.NewAPI(cfg)
	api.ServeHttp(cfg.API_PORT)
}
