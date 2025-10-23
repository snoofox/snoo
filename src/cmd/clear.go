package cmd

import (
	"fmt"
	"snoo/src/db"

	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear cached posts and comments to force fresh fetch",
	Run: func(cmd *cobra.Command, args []string) {
		database := db.FromContext(cmd.Context())

		result := database.Unscoped().Delete(&db.Post{}, "1=1")
		if result.Error != nil {
			fmt.Printf("Error clearing posts cache: %v\n", result.Error)
			return
		}
		postsDeleted := result.RowsAffected

		result = database.Unscoped().Delete(&db.Comment{}, "1=1")
		if result.Error != nil {
			fmt.Printf("Error clearing comments cache: %v\n", result.Error)
			return
		}
		commentsDeleted := result.RowsAffected

		result = database.Model(&db.Source{}).Where("1=1").Update("last_fetch_at", nil)
		if result.Error != nil {
			fmt.Printf("Error resetting source fetch times: %v\n", result.Error)
			return
		}

		fmt.Printf("Cache cleared successfully!\n")
		fmt.Printf("- Deleted %d posts\n", postsDeleted)
		fmt.Printf("- Deleted %d comments\n", commentsDeleted)
		fmt.Printf("- Reset fetch times for all sources\n")
		fmt.Printf("\nNext feed fetch will get fresh data.\n")
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
