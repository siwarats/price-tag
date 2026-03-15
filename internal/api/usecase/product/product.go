package usecase

import (
	"github.com/siwarats/price-tag/internal/api/entity"
	repository "github.com/siwarats/price-tag/internal/api/repository/product"
)

type ProductUseCase interface {
	GetProduct(id string) (*entity.Product, error)
	GetProducts() ([]entity.Product, error)
}

type productUseCase struct {
	repository repository.ProductRepository
}

func NewProductUseCase(repository repository.ProductRepository) ProductUseCase {
	return &productUseCase{
		repository: repository,
	}
}
