package repository

import (
	"context"
	"Permia/core-service/internal/domain"

	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// GetActiveProducts گرفتن لیست محصولات فعال مرتب شده
func (r *productRepository) GetActiveProducts(ctx context.Context) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("display_order asc").
		Find(&products).Error
	return products, err
}

func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&product).Error
	return &product, err
}