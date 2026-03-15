package usecase

import (
	"github.com/siwarats/price-tag/internal/api/entity"
)

func (u *productUseCase) GetProducts() ([]entity.Product, error) {
	return u.repository.GetProducts()
}
