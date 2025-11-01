package db

import (
	"context"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type contextKey string

const dbKey contextKey = "db"

func GetDB() (*gorm.DB, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Source{}, &Post{}, &Comment{}, &Setting{})

	return db, nil
}

func getDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	snooDir := filepath.Join(homeDir, ".snoo")
	if err := os.MkdirAll(snooDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(snooDir, "data.sqlite3"), nil
}

func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}

func FromContext(ctx context.Context) *gorm.DB {
	return ctx.Value(dbKey).(*gorm.DB)
}
