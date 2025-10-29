package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show help and navigation keys",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(`
snoo - A terminal feed reader that doesn't suck (yet).

COMMANDS:
  snoo                       Open feed (default)
  snoo sub add <name>        Subscribe to a subreddit
  snoo sub rss <url>         Subscribe to an RSS feed
  snoo sub lobsters <cat>    Subscribe to Lobsters (active or recent)
  snoo sub hn <cat>          Subscribe to HackerNews (top, new, best, ask, show, job)
  snoo sub list              List all subscriptions
  snoo sub rm <id>           Remove a subscription
  snoo theme <name>          Change theme (default, catppuccin, dracula, github, peppermint)
  snoo clear                 Clear all data
  snoo help                  Show this help

NAVIGATION KEYS:

Feed List View:
  j / ↓         Move down
  k / ↑         Move up
  Enter/Space   Open selected post
  s             Sort posts (by upvotes, comments, date)
  f             Filter sources (toggle subscriptions on/off)
  q             Quit

Post View:
  j / ↓         Scroll down
  k / ↑         Scroll up
  g             Go to top
  G             Go to bottom
  r             Read full article (toggle between original and article)
  s             Sort comments (by score, date)
  Esc/Backspace/q Back to feed list
  q             Back to feed list

Sort Menu:
  j / ↓         Move down
  k / ↑         Move up
  Enter/Space   Select sort option
  Esc/Backspace Cancel

Filter Menu:
  j / ↓         Move down
  k / ↑         Move up
  Space         Toggle source on/off
  a             Enable all sources
  d             Disable all sources
  Esc/Backspace Back to feed list
`)
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
