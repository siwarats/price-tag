package main

import (
	"github.com/siwarats/price-tag/internal/scraper/lotuss"
	"github.com/siwarats/price-tag/internal/shared"
)

func main() {
	cfg := shared.NewConfig()
	l := lotuss.NewLotuss(cfg)
	l.RunAPI(cfg.SCRAPER_LOTUSS_PORT)
}
