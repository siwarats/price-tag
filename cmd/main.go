package main

import (
	"github.com/siwarats/price-tag/internal/scraper/lotuss"
	"github.com/siwarats/price-tag/pkg"
)

func main() {
	cfg := pkg.NewConfig()

	lotuss := lotuss.NewLotuss(cfg)
	lotuss.Run()
}
