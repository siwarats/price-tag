package repository

import "github.com/siwarats/price-tag/internal/api/entity"

type ProductRepository interface {
	GetProduct(id string) (*entity.Product, error)
	GetProducts() ([]entity.Product, error)
}

type productRepository struct {
}

func NewProductRepository() ProductRepository {
	return &productRepository{}
}
