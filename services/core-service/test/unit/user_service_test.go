package unit

import (
	"context"
	"testing"

	"Permia/core-service/internal/domain"
	"Permia/core-service/internal/service"
	"Permia/core-service/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOrCreateUser_NewUser(t *testing.T) {
	// 1. ساخت ماک و سرویس
	mockRepo := new(mocks.MockUserRepository)
	userService := service.NewUserService(mockRepo)

	// 2. سناریو: وقتی گت زدی، بگو "کاربر نیست" (nil)
	telegramID := int64(123456)
	mockRepo.On("GetByTelegramID", mock.Anything, telegramID).Return(nil, nil)

	// 3. سناریو: وقتی کریت زدی، بگو "موفق شد" (nil error)
	// در اینجا چک می‌کنیم که کد ریفرال ساخته شده باشد
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.TelegramID == telegramID && 
		       u.Username == "test_user" && 
		       u.ReferralCode != "" // مطمئن می‌شویم کد دعوت تولید شده
	})).Return(nil)

	// 4. اجرای تابع واقعی (اصلاح شده با پارامتر جدید)
	// پارامتر آخر ("") همان کد معرف است که اینجا خالی رد می‌کنیم
	user, err := userService.GetOrCreateUser(context.Background(), telegramID, "test_user", "Test", "User", "")

	// 5. بررسی نتایج (Assertions)
	assert.NoError(t, err)           // نباید ارور بدهد
	assert.NotNil(t, user)           // یوزر نباید خالی باشد
	assert.Equal(t, telegramID, user.TelegramID)
	assert.Equal(t, "test_user", user.Username)
	assert.NotEmpty(t, user.ReferralCode) // چک کردن تولید کد دعوت

	// مطمئن می‌شویم که متدهای دیتابیس واقعا صدا زده شدند
	mockRepo.AssertExpectations(t)
}