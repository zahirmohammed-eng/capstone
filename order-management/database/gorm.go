package database

import (
	"log"
	"time"

	"github.com/glebarez/sqlite" // Pure Go SQLite (zero CGO required)
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitGormDB initializes GORM with connection pooling.
// If PostgreSQL is unavailable, it automatically falls back to pure-Go SQLite
// so the microservices run immediately on Windows without GCC or CGO.
func InitGormDB(dsn string, fallbackSQLiteFile string) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	// 1. Try connecting to PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		log.Printf(" PostgreSQL not reachable (%v). Falling back to pure-Go SQLite: %s", err, fallbackSQLiteFile)
		
		// 2. Fallback to Pure-Go SQLite (works with CGO_ENABLED=0)
		db, err = gorm.Open(sqlite.Open(fallbackSQLiteFile), gormCfg)
		if err != nil {
			return nil, err
		}
	} else {
		log.Printf(" Connected successfully to PostgreSQL: %s", dsn)
	}

	// 3. Configure Connection Pooling
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetConnMaxLifetime(15 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	}

	return db, nil
}