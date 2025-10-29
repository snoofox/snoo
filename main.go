package main

import (
	"context"
	"log"

	"github.com/snoofox/snoo/src/cmd"
	"github.com/snoofox/snoo/src/db"
	"github.com/snoofox/snoo/src/feed"
	"github.com/snoofox/snoo/src/providers/hackernews"
	"github.com/snoofox/snoo/src/providers/lobsters"
	"github.com/snoofox/snoo/src/providers/reddit"
	"github.com/snoofox/snoo/src/providers/rss"
)

func main() {
	feed.Register(reddit.New())
	feed.Register(rss.New())
	feed.Register(lobsters.New())
	feed.Register(hackernews.New())

	database, err := db.GetDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	ctx := db.WithDB(context.Background(), database)
	cmd.Execute(ctx)
}
