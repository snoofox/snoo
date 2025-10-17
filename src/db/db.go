package db

import (
	"context"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type contextKey string

const dbKey contextKey = "db"

func GetDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("data.sqlite3"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Subreddit{}, &Post{}, &Comment{})

	return db, nil
}

func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}

func FromContext(ctx context.Context) *gorm.DB {
	return ctx.Value(dbKey).(*gorm.DB)
}
