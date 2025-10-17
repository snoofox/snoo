package cmd

import (
	"fmt"
	"snoo/src/db"
	"snoo/src/reddit"
	"strconv"

	"github.com/spf13/cobra"
)

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Subscribe / unsubscribe / list subs",
}

var subListCmd = &cobra.Command{
	Use:   "list",
	Short: "List subscribed subreddits",
	Run: func(cmd *cobra.Command, args []string) {
		database := db.FromContext(cmd.Context())

		var subreddits []db.Subreddit
		result := database.Find(&subreddits)

		if result.Error != nil {
			fmt.Printf("Error fetching subreddits: %v\n", result.Error)
			return
		}

		if len(subreddits) == 0 {
			fmt.Println("No subscribed subreddits")
			return
		}

		fmt.Printf("Subscribed to %d subreddit(s):\n\n", len(subreddits))
		for _, sub := range subreddits {
			fmt.Printf("%d. r/%s\n", sub.ID, sub.DisplayName)
			if sub.Desc != "" {
				fmt.Printf("  %s\n", sub.Desc)
			}
			fmt.Printf("  Subscribers: %.0f\n\n", sub.Subscribers)
		}
	},
}

var subAddCmd = &cobra.Command{
	Use:   "add SUBREDDIT",
	Short: "Subscribe to a subreddit",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		database := db.FromContext(ctx)
		resp, err := reddit.FetchSubreddit(args[0])
		if err != nil {
			fmt.Printf("Error fetching subreddit: %s\n", args[0])
			return
		}

		subreddit := &db.Subreddit{
			RedditID:    resp.RedditID,
			Name:        resp.Name,
			DisplayName: resp.DisplayName,
			Desc:        resp.Desc,
			Subscribers: resp.Subscribers,
			NSFW:        resp.NSFW,
			IconURL:     resp.IconURL,
			CreatedUTC:  resp.CreatedUTC,
			LastFetchAt: resp.LastFetchAt,
		}

		result := database.Create(subreddit)
		if result.Error != nil {
			// Check for duplicate key error
			if result.Error.Error() == "UNIQUE constraint failed: subreddits.reddit_id" ||
				result.Error.Error() == "UNIQUE constraint failed: subreddits.name" {
				fmt.Printf("Already subscribed to r/%s\n", args[0])
				return
			}
			fmt.Printf("Error subscribing to r/%s: %v\n", args[0], result.Error)
			return
		}

		fmt.Printf("Successfully subscribed to r/%s\n", resp.DisplayName)
	},
}

var subRmCmd = &cobra.Command{
	Use:   "rm ID",
	Short: "Unsubscribe from a subreddit",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Please provide a numeric ID")
			return
		}
		database := db.FromContext(cmd.Context())

		result := database.Unscoped().Delete(&db.Subreddit{}, id)
		if result.Error != nil {
			fmt.Printf("Error unsubscribing: %v\n", result.Error)
			return
		}

		if result.RowsAffected == 0 {
			fmt.Printf("No subreddit found with ID %d\n", id)
			return
		}

		fmt.Println("Successfully unsubscribed")
	},
}

func init() {
	rootCmd.AddCommand(subCmd)
	subCmd.AddCommand(subListCmd, subAddCmd, subRmCmd)
}
