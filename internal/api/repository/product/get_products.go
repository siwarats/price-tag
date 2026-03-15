package repository

import (
	"github.com/siwarats/price-tag/internal/api/entity"
)

func (r *productRepository) GetProducts() ([]entity.Product, error) {
	return []entity.Product{}, nil
}