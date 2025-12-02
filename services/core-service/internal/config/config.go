package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Config ساختار کلی تنظیمات برنامه
type Config struct {
	AppEnv     string
	ServerPort string
	Database   DatabaseConfig
	Brocard    BrocardConfig
	Zarinpal   ZarinpalConfig
	Security   SecurityConfig
}

// DatabaseConfig تنظیمات دیتابیس
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// BrocardConfig تنظیمات سرویس‌دهنده کارت مجازی
type BrocardConfig struct {
	APIToken string
}

// ZarinpalConfig تنظیمات درگاه پرداخت زرین‌پال
type ZarinpalConfig struct {
	MerchantID string
}

// SecurityConfig تنظیمات امنیتی
type SecurityConfig struct {
	AdminIDs []string
}

// Load تلاش می‌کند فایل .env را از پوشه deployment پیدا و لود کند
func Load() (*Config, error) {
	// 1. پیدا کردن و لود کردن فایل .env
	envPath, err := findDeploymentEnv()
	if err == nil {
		// اگر فایل پیدا شد، آن را لود کن
		if loadErr := godotenv.Load(envPath); loadErr != nil {
			return nil, fmt.Errorf("error loading .env file: %v", loadErr)
		}
		fmt.Printf("✅ Loaded configuration from: %s\n", envPath)
	} else {
		// اگر فایل نبود، شاید در پروداکشن هستیم و متغیرها ست شده‌اند
		fmt.Println("⚠️  Warning: .env file not found in deployment folder, using existing system environment variables.")
	}

	// 2. پر کردن استراکت کانفیگ (از متغیرهای محیطی که الان ست شده‌اند)
	cfg := &Config{
		AppEnv:     getEnv("APP_ENV", "development"),
		ServerPort: getEnv("CORE_PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "permia_db"),
		},
		Brocard: BrocardConfig{
			APIToken: getEnv("BROCARD_API_TOKEN", ""),
		},
		Zarinpal: ZarinpalConfig{
			MerchantID: getEnv("ZARINPAL_MERCHANT_ID", ""),
		},
		Security: SecurityConfig{
			AdminIDs: strings.Split(getEnv("ADMIN_IDS", ""), ","),
		},
	}

	return cfg, nil
}

// findDeploymentEnv به دنبال فایل deployment/.env می‌گردد
func findDeploymentEnv() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// تا 5 سطح به عقب برمی‌گردیم تا پوشه deployment را پیدا کنیم
	for i := 0; i < 5; i++ {
		// مسیر احتمالی: current_dir/deployment/.env
		path := filepath.Join(dir, "deployment", ".env")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		// مسیر احتمالی دوم: شاید فایل .env مستقیم در روت باشد (برای سازگاری)
		rootPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(rootPath); err == nil {
			return rootPath, nil
		}

		// رفتن به دایرکتوری والد
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find deployment/.env file")
}

// getEnv یک تابع کمکی برای خواندن متغیر با مقدار پیش‌فرض
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
