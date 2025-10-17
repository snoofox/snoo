package main

import (
	"context"
	"log"
	"snoo/src/cmd"
	"snoo/src/db"
	"snoo/src/reddit"
)

func main() {
	database, err := db.GetDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	ctx := db.WithDB(context.Background(), database)

	reddit.StartCleanupTicker(ctx)

	cmd.Execute(ctx)
}
