package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "snoo",
	Short: "reddit reader",
	Run: func(cmd *cobra.Command, args []string) {
		feedCmd.Run(cmd, args)
	},
}

func Execute(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
