package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "snoo",
	Short: "A terminal feed reader that doesn't suck (yet)",
	Long:  "snoo - A terminal feed reader that doesn't suck (yet).\n\nA fast, keyboard-driven feed reader for Reddit, RSS, Lobsters, and Hacker News.",
	Run: func(cmd *cobra.Command, args []string) {
		feedCmd.Run(cmd, args)
	},
}

func Execute(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
