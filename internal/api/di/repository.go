package di

import (
	product "github.com/siwarats/price-tag/internal/api/repository/product"
)

type repository struct {
	product product.ProductRepository
}

func NewRepository() *repository {
	return &repository{
		product: product.NewProductRepository(),
	}
}
