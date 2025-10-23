package main

import (
	"context"
	"log"
	"snoo/src/cmd"
	"snoo/src/db"
	"snoo/src/feed"
	"snoo/src/providers/reddit"
	"snoo/src/providers/rss"
)

func main() {
	feed.Register(reddit.New())
	feed.Register(rss.New())

	database, err := db.GetDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	ctx := db.WithDB(context.Background(), database)
	cmd.Execute(ctx)
}
