package service

import (
	"Permia/core-service/internal/domain"
	"context"
)

type ProductService struct {
	repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// GetCatalog لیست محصولات دسته‌بندی شده
func (s *ProductService) GetCatalog(ctx context.Context) (map[string][]domain.Product, error) {
	products, err := s.repo.GetActiveProducts(ctx)
	if err != nil {
		return nil, err
	}

	// دسته‌بندی محصولات برای نمایش بهتر در بات
	catalog := make(map[string][]domain.Product)
	for _, p := range products {
		catalog[p.Category] = append(catalog[p.Category], p)
	}

	return catalog, nil
}

// GetProductBySKU دریافت محصول برحسب SKU
func (s *ProductService) GetProductBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	return s.repo.GetBySKU(ctx, sku)
}
