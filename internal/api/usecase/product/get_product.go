package usecase

import (
	"github.com/siwarats/price-tag/internal/api/entity"
)

func (u *productUseCase) GetProduct(id string) (*entity.Product, error) {
	return u.repository.GetProduct(id)
}
