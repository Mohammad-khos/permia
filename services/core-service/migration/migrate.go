package migration

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

// Up فقط ساختار جداول را مدیریت می‌کند
func Up(db *gorm.DB) {
	log.Println("🔄 Starting Database Schema Migration...")

	runSQLMigrations(db)
	
	// دیگر نیازی به seedProducts نیست چون دستی در pgAdmin زدید
}

func runSQLMigrations(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get sql.DB: %v", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to create migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("❌ Migration init failed: %v", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("✅ Schema is up to date.")
		} else {
			log.Fatalf("❌ Migration UP failed: %v", err)
		}
	} else {
		log.Println("✅ SQL Schema Migrated Successfully.")
	}
}