package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/snoofox/snoo/src/db"
	"github.com/spf13/cobra"
)

type Theme struct {
	Name       string
	Title      lipgloss.Color
	Selected   lipgloss.Color
	Cursor     lipgloss.Color
	Subreddit  lipgloss.Color
	Score      lipgloss.Color
	NSFWBg     lipgloss.Color
	NSFWFg     lipgloss.Color
	Dim        lipgloss.Color
	URL        lipgloss.Color
	Comments   lipgloss.Color
	Separator  lipgloss.Color
	HelpNav    lipgloss.Color
	HelpAction lipgloss.Color
	HelpQuit   lipgloss.Color
}

var themes = map[string]Theme{
	"default": {
		Name:       "Default",
		Title:      lipgloss.Color("#FF6B9D"),
		Selected:   lipgloss.Color("#FAFAFA"),
		Cursor:     lipgloss.Color("#F472B6"),
		Subreddit:  lipgloss.Color("#7C9CBF"),
		Score:      lipgloss.Color("#86EFAC"),
		NSFWBg:     lipgloss.Color("#F87171"),
		NSFWFg:     lipgloss.Color("#1A1A1A"),
		Dim:        lipgloss.Color("#6B7280"),
		URL:        lipgloss.Color("#818CF8"),
		Comments:   lipgloss.Color("#D4A574"),
		Separator:  lipgloss.Color("#3F3F46"),
		HelpNav:    lipgloss.Color("#60A5FA"),
		HelpAction: lipgloss.Color("#34D399"),
		HelpQuit:   lipgloss.Color("#F87171"),
	},
	"catppuccin": {
		Name:       "Catppuccin Mocha",
		Title:      lipgloss.Color("#F5C2E7"),
		Selected:   lipgloss.Color("#CDD6F4"),
		Cursor:     lipgloss.Color("#F38BA8"),
		Subreddit:  lipgloss.Color("#89B4FA"),
		Score:      lipgloss.Color("#A6E3A1"),
		NSFWBg:     lipgloss.Color("#F38BA8"),
		NSFWFg:     lipgloss.Color("#1E1E2E"),
		Dim:        lipgloss.Color("#6C7086"),
		URL:        lipgloss.Color("#B4BEFE"),
		Comments:   lipgloss.Color("#FAB387"),
		Separator:  lipgloss.Color("#45475A"),
		HelpNav:    lipgloss.Color("#89DCEB"),
		HelpAction: lipgloss.Color("#A6E3A1"),
		HelpQuit:   lipgloss.Color("#F38BA8"),
	},
	"dracula": {
		Name:       "Dracula",
		Title:      lipgloss.Color("#FF79C6"),
		Selected:   lipgloss.Color("#F8F8F2"),
		Cursor:     lipgloss.Color("#FF79C6"),
		Subreddit:  lipgloss.Color("#8BE9FD"),
		Score:      lipgloss.Color("#50FA7B"),
		NSFWBg:     lipgloss.Color("#FF5555"),
		NSFWFg:     lipgloss.Color("#282A36"),
		Dim:        lipgloss.Color("#6272A4"),
		URL:        lipgloss.Color("#BD93F9"),
		Comments:   lipgloss.Color("#FFB86C"),
		Separator:  lipgloss.Color("#44475A"),
		HelpNav:    lipgloss.Color("#8BE9FD"),
		HelpAction: lipgloss.Color("#50FA7B"),
		HelpQuit:   lipgloss.Color("#FF5555"),
	},
	"github": {
		Name:       "GitHub",
		Title:      lipgloss.Color("#1F6FEB"),
		Selected:   lipgloss.Color("#24292F"),
		Cursor:     lipgloss.Color("#0969DA"),
		Subreddit:  lipgloss.Color("#8250DF"),
		Score:      lipgloss.Color("#1A7F37"),
		NSFWBg:     lipgloss.Color("#CF222E"),
		NSFWFg:     lipgloss.Color("#FFFFFF"),
		Dim:        lipgloss.Color("#656D76"),
		URL:        lipgloss.Color("#0969DA"),
		Comments:   lipgloss.Color("#953800"),
		Separator:  lipgloss.Color("#D0D7DE"),
		HelpNav:    lipgloss.Color("#0969DA"),
		HelpAction: lipgloss.Color("#1A7F37"),
		HelpQuit:   lipgloss.Color("#CF222E"),
	},
	"peppermint": {
		Name:       "Peppermint",
		Title:      lipgloss.Color("#00D9A3"),
		Selected:   lipgloss.Color("#E8FFF8"),
		Cursor:     lipgloss.Color("#00FFC2"),
		Subreddit:  lipgloss.Color("#00B8D4"),
		Score:      lipgloss.Color("#76FF03"),
		NSFWBg:     lipgloss.Color("#FF1744"),
		NSFWFg:     lipgloss.Color("#FFFFFF"),
		Dim:        lipgloss.Color("#546E7A"),
		URL:        lipgloss.Color("#00E5FF"),
		Comments:   lipgloss.Color("#FFD740"),
		Separator:  lipgloss.Color("#37474F"),
		HelpNav:    lipgloss.Color("#00E5FF"),
		HelpAction: lipgloss.Color("#76FF03"),
		HelpQuit:   lipgloss.Color("#FF1744"),
	},
}

var currentTheme = "default"

func GetCurrentTheme() Theme {
	if theme, ok := themes[currentTheme]; ok {
		return theme
	}
	return themes["default"]
}

func SetTheme(name string) bool {
	if _, ok := themes[name]; ok {
		currentTheme = name
		applyTheme()
		return true
	}
	return false
}

func GetCurrentThemeName() string {
	return currentTheme
}

func GetThemeNames() []string {
	names := make([]string, 0, len(themes))
	for name := range themes {
		names = append(names, name)
	}
	return names
}

func applyTheme() {
	theme := GetCurrentTheme()

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Title)

	selectedStyle = lipgloss.NewStyle().
		Foreground(theme.Selected).
		Bold(true)

	cursorStyle = lipgloss.NewStyle().
		Foreground(theme.Cursor).
		Bold(true)

	subredditStyle = lipgloss.NewStyle().
		Foreground(theme.Subreddit)

	scoreStyle = lipgloss.NewStyle().
		Foreground(theme.Score)

	nsfwStyle = lipgloss.NewStyle().
		Foreground(theme.NSFWFg).
		Background(theme.NSFWBg).
		Bold(true).
		Padding(0, 1)

	dimStyle = lipgloss.NewStyle().
		Foreground(theme.Dim)

	urlStyle = lipgloss.NewStyle().
		Foreground(theme.URL)

	commentsStyle = lipgloss.NewStyle().
		Foreground(theme.Comments)

	separatorStyle = lipgloss.NewStyle().
		Foreground(theme.Separator)
}

var themeCmd = &cobra.Command{
	Use:   "theme [name]",
	Short: "Change the color theme",
	Long:  "Change the color theme for the feed viewer. Available themes: default, catppuccin, dracula, github, peppermint",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			listThemes()
			return
		}

		themeName := args[0]
		if SetTheme(themeName) {
			if gormDB := db.FromContext(cmd.Context()); gormDB != nil {
				db.SetSetting(gormDB, "theme", themeName)
			}
			fmt.Printf("✓ Theme changed to: %s\n", themeName)
		} else {
			fmt.Printf("✗ Unknown theme: %s\n", themeName)
			fmt.Println("\nAvailable themes:")
			listThemes()
		}
	},
}

func listThemes() {
	fmt.Println("Available themes:")

	themeNames := []string{"default", "catppuccin", "dracula", "github", "peppermint"}

	for _, name := range themeNames {
		theme := themes[name]
		current := ""
		if name == currentTheme {
			current = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D9A3")).
				Bold(true).
				Render(" (current)")
		}

		titleStyle := lipgloss.NewStyle().
			Foreground(theme.Title).
			Bold(true)

		fmt.Printf("  %s%s\n", titleStyle.Render(name), current)

		preview := fmt.Sprintf("    Preview: %s %s %s %s %s",
			lipgloss.NewStyle().Foreground(theme.Subreddit).Render("subreddit"),
			lipgloss.NewStyle().Foreground(theme.Score).Render("↑123"),
			lipgloss.NewStyle().Foreground(theme.Comments).Render("󰆉 45"),
			lipgloss.NewStyle().Foreground(theme.URL).Render("link"),
			lipgloss.NewStyle().
				Foreground(theme.NSFWFg).
				Background(theme.NSFWBg).
				Padding(0, 1).
				Render("NSFW"),
		)

		fmt.Println(preview)
		fmt.Println()
	}

	dimPreview := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	fmt.Println(dimPreview.Render("Usage: snoo theme [name]"))
}

func init() {
	rootCmd.AddCommand(themeCmd)
}
