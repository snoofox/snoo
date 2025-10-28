package cmd

import (
	"fmt"
	"strconv"

	"github.com/snoofox/snoo/src/db"
	"github.com/snoofox/snoo/src/feed"
	"github.com/spf13/cobra"
)

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Manage subscriptions",
}

var subListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all subscribed sources",
	Run: func(cmd *cobra.Command, args []string) {
		database := db.FromContext(cmd.Context())
		manager := feed.NewManager(database)

		sources, err := manager.ListSources()
		if err != nil {
			fmt.Printf("Error listing sources: %v\n", err)
			return
		}

		if len(sources) == 0 {
			fmt.Println("No subscribed sources")
			fmt.Println("\nAvailable commands:")
			fmt.Println("  snoo sub add <subreddit:sort>  - Subscribe to a subreddit (hot, new, rising, top, best)")
			fmt.Println("  snoo sub rss <url>               - Subscribe to an RSS feed")
			fmt.Println("  snoo sub lobsters <category>     - Subscribe to Lobsters (active or recent)")
			return
		}

		fmt.Printf("Subscribed to %d source(s):\n\n", len(sources))
		for _, src := range sources {
			fmt.Printf("%d. [%s] %s\n", src.ID, src.Type, src.DisplayName)
			if src.Description != "" {
				fmt.Printf("   %s\n", src.Description)
			}
			fmt.Printf("   Identifier: %s\n\n", src.Identifier)
		}
	},
}

var subAddCmd = &cobra.Command{
	Use:   "add SUBREDDIT[:SORT]",
	Short: "Subscribe to a subreddit (sort: hot, new, rising, top, best)",
	Long: `Subscribe to a subreddit with optional sort type.

Examples:
  snoo sub add golang          # defaults to 'best'
  snoo sub add golang:hot      # hot posts
  snoo sub add golang:new      # new posts
  snoo sub add golang:rising   # rising posts
  snoo sub add golang:top      # top posts`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		database := db.FromContext(ctx)
		manager := feed.NewManager(database)

		identifier := args[0]
		fmt.Printf("Subscribing to r/%s...\n", identifier)
		if err := manager.Subscribe(ctx, "reddit", identifier); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Successfully subscribed to r/%s\n", identifier)
	},
}

var rssAddCmd = &cobra.Command{
	Use:   "rss URL",
	Short: "Subscribe to an RSS feed",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		database := db.FromContext(ctx)
		manager := feed.NewManager(database)

		fmt.Printf("Subscribing to RSS feed...\n")
		if err := manager.Subscribe(ctx, "rss", args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("Successfully subscribed to RSS feed")
	},
}

var lobstersAddCmd = &cobra.Command{
	Use:   "lobsters CATEGORY",
	Short: "Subscribe to Lobsters (active or recent)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		database := db.FromContext(ctx)
		manager := feed.NewManager(database)

		category := args[0]
		if category != "active" && category != "recent" {
			fmt.Println("Error: category must be 'active' or 'recent'")
			return
		}

		fmt.Printf("Subscribing to Lobsters %s...\n", category)
		if err := manager.Subscribe(ctx, "lobsters", category); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Successfully subscribed to Lobsters %s\n", category)
	},
}

var subRmCmd = &cobra.Command{
	Use:   "rm ID",
	Short: "Unsubscribe from a source",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Println("Please provide a numeric ID")
			return
		}

		database := db.FromContext(cmd.Context())
		manager := feed.NewManager(database)

		if err := manager.Unsubscribe(uint(id)); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("Successfully unsubscribed")
	},
}

func init() {
	rootCmd.AddCommand(subCmd)
	subCmd.AddCommand(subListCmd, subAddCmd, rssAddCmd, lobstersAddCmd, subRmCmd)
}
