package repository

import (
	"context"
	"errors"
	"Permia/core-service/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) domain.AccountRepository {
	return &accountRepository{db: db}
}

// GetAvailableAccount الگوریتم هوشمند پیدا کردن اکانت (نیمه‌پر یا خالی)
func (r *accountRepository) GetAvailableAccount(ctx context.Context, productSKU string) (*domain.AccountInventory, error) {
	var account domain.AccountInventory

	// گام ۱: تلاش برای پیدا کردن اکانت "نیمه‌پر" (برای پلن‌های اشتراکی)
	// نکته: Locking (FOR UPDATE) استفاده شده تا دو نفر همزمان یک صندلی را نگیرند
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("product_sku = ? AND status = ? AND current_users > 0 AND current_users < max_users", productSKU, "AVAILABLE").
		First(&account).Error

	if err == nil {
		return &account, nil // اکانت نیمه‌پر پیدا شد
	}

	// گام ۲: اگر اکانت نیمه‌پر نبود، دنبال اکانت "کاملاً خالی" بگرد
	err = r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("product_sku = ? AND status = ? AND current_users = 0", productSKU, "AVAILABLE").
		First(&account).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("out of stock") // انبار خالی است
		}
		return nil, err
	}

	return &account, nil
}

// MarkAsSold افزایش تعداد استفاده‌کنندگان و تغییر وضعیت در صورت پر شدن
func (r *accountRepository) MarkAsSold(ctx context.Context, accountID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var acc domain.AccountInventory
		if err := tx.First(&acc, accountID).Error; err != nil {
			return err
		}

		// یکی به تعداد یوزرها اضافه کن
		acc.CurrentUsers++

		// اگر ظرفیت پر شد، وضعیت را تغییر بده
		if acc.CurrentUsers >= acc.MaxUsers {
			acc.Status = "FILLED"
		}

		return tx.Save(&acc).Error
	})
}

// CreateBatch اضافه کردن دستی اکانت‌ها توسط ادمین
func (r *accountRepository) CreateBatch(ctx context.Context, accounts []domain.AccountInventory) error {
	return r.db.WithContext(ctx).Create(&accounts).Error
}