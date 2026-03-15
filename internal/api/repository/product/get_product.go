package repository

import (
	"github.com/siwarats/price-tag/internal/api/entity"
)

func (r *productRepository) GetProduct(id string) (*entity.Product, error) {
	return &entity.Product{}, nil
}
