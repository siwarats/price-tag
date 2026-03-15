package di

import (
	product "github.com/siwarats/price-tag/internal/api/usecase/product"
)

type useCase struct {
	product product.ProductUseCase
}

func NewUseCase(
	repository *repository,
) *useCase {
	return &useCase{
		product: product.NewProductUseCase(repository.product),
	}
}
