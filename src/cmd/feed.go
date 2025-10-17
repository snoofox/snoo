package cmd

import (
	"context"
	"fmt"
	"snoo/src/db"
	"snoo/src/reddit"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	titleStyle     lipgloss.Style
	selectedStyle  lipgloss.Style
	cursorStyle    lipgloss.Style
	subredditStyle lipgloss.Style
	scoreStyle     lipgloss.Style
	nsfwStyle      lipgloss.Style
	dimStyle       lipgloss.Style
	urlStyle       lipgloss.Style
	commentsStyle  lipgloss.Style
	separatorStyle lipgloss.Style
)

type commentsLoadedMsg struct {
	comments []reddit.Comment
}

type model struct {
	posts           []reddit.Post
	cursor          int
	viewing         bool
	selected        int
	width           int
	height          int
	comments        []reddit.Comment
	loadingComments bool
	ctx             context.Context
	viewport        viewport.Model
	listViewport    viewport.Model
	ready           bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-2)
			m.listViewport = viewport.New(msg.Width, msg.Height-2)
			m.listViewport.SetContent(m.renderListContent())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 2
			m.listViewport.Width = msg.Width
			m.listViewport.Height = msg.Height - 2
			m.listViewport.SetContent(m.renderListContent())
		}

	case commentsLoadedMsg:
		m.comments = msg.comments
		m.loadingComments = false
		m.viewport.SetContent(m.renderPostContent())
		m.viewport.GotoTop()

	case tea.KeyMsg:
		if m.viewing {
			switch msg.String() {
			case "q", "esc", "backspace":
				m.viewing = false
				m.comments = nil
				m.loadingComments = false
				return m, nil
			case "up", "k":
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			case "down", "j":
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		} else {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					m.listViewport.SetContent(m.renderListContent())
					m.ensureCursorVisible()
				}
			case "down", "j":
				if m.cursor < len(m.posts)-1 {
					m.cursor++
					m.listViewport.SetContent(m.renderListContent())
					m.ensureCursorVisible()
				}
			case "enter", " ":
				m.selected = m.cursor
				m.viewing = true
				m.loadingComments = true
				m.comments = nil
				m.viewport.SetContent(m.renderPostContent())
				m.viewport.GotoTop()
				return m, m.loadCommentsCmd()
			}
		}
	}

	if m.viewing {
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.viewing {
		return m.viewPost()
	}
	return m.viewList()
}

func (m *model) ensureCursorVisible() {
	headerLines := 4
	linesPerPost := 3
	cursorLine := headerLines + (m.cursor * linesPerPost)

	viewportTop := m.listViewport.YOffset
	viewportBottom := viewportTop + m.listViewport.Height

	if cursorLine < viewportTop {
		m.listViewport.SetYOffset(cursorLine)
	} else if cursorLine+linesPerPost > viewportBottom {
		newOffset := cursorLine - m.listViewport.Height + linesPerPost
		if newOffset < 0 {
			newOffset = 0
		}
		m.listViewport.SetYOffset(newOffset)
	}
}

func (m model) viewList() string {
	if !m.ready {
		return "Loading..."
	}
	theme := GetCurrentTheme()
	helpText := dimStyle.Render("  ") +
		lipgloss.NewStyle().Foreground(theme.HelpNav).Render("j/k") +
		dimStyle.Render(" navigate  ") +
		lipgloss.NewStyle().Foreground(theme.HelpAction).Render("enter") +
		dimStyle.Render(" read  ") +
		lipgloss.NewStyle().Foreground(theme.HelpQuit).Render("q") +
		dimStyle.Render(" quit")
	return m.listViewport.View() + "\n" + helpText
}

func (m model) renderListContent() string {
	s := "\n"
	s += titleStyle.Render("  󰑍  Your Feed") + "\n"
	s += dimStyle.Render(fmt.Sprintf("  %d posts", len(m.posts))) + "\n\n"

	for i, post := range m.posts {
		titleText := truncate(post.Title, 85)

		if m.cursor == i {
			cursor := cursorStyle.Render("● ")

			nsfw := ""
			if post.NSFW {
				nsfw = nsfwStyle.Render(" NSFW ") + " "
			}

			sub := subredditStyle.Render(post.Subreddit)
			score := scoreStyle.Render(fmt.Sprintf(" %d", post.Score))
			comments := commentsStyle.Render(fmt.Sprintf("󰆉 %d", post.NumComments))
			sep := separatorStyle.Render("•")

			s += " " + cursor + selectedStyle.Render(titleText) + "\n"
			s += "   " + nsfw + sub + " " + sep + " " + score + " " + sep + " " + comments + "\n\n"
		} else {
			nsfw := ""
			if post.NSFW {
				nsfw = nsfwStyle.Render(" NSFW ") + " "
			}

			sub := subredditStyle.Render(post.Subreddit)
			score := scoreStyle.Render(fmt.Sprintf(" %d", post.Score))
			comments := commentsStyle.Render(fmt.Sprintf("󰆉 %d", post.NumComments))
			sep := separatorStyle.Render("•")

			s += "   " + lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB")).Render(titleText) + "\n"
			s += "   " + nsfw + sub + " " + sep + " " + score + " " + sep + " " + comments + "\n\n"
		}
	}

	return s
}

func (m model) loadCommentsCmd() tea.Cmd {
	return func() tea.Msg {
		post := m.posts[m.selected]
		comments, err := reddit.FetchComments(m.ctx, post.Permalink)
		if err != nil {
			return commentsLoadedMsg{comments: []reddit.Comment{}}
		}
		return commentsLoadedMsg{comments: comments}
	}
}

func (m model) renderPostContent() string {
	post := m.posts[m.selected]
	maxWidth := m.width - 4
	if maxWidth < 40 {
		maxWidth = 40
	}

	s := "\n" + titleStyle.Render(wrapText(post.Title, maxWidth)) + "\n\n"

	if post.IsSelf && post.Selftext != "" {
		s += wrapText(post.Selftext, maxWidth) + "\n"
	} else if !post.IsSelf {
		s += urlStyle.Render(post.URL) + "\n"
	}

	s += "\n" + dimStyle.Render(fmt.Sprintf("─── %d comments ───", post.NumComments)) + "\n\n"

	if m.loadingComments {
		s += dimStyle.Render("Loading comments...") + "\n"
	} else if len(m.comments) == 0 {
		s += dimStyle.Render("No comments yet") + "\n"
	} else {
		for _, comment := range m.comments {
			s += renderComment(comment, maxWidth) + "\n"
		}
	}

	return s
}

func (m model) viewPost() string {
	return m.viewport.View() + "\n" + dimStyle.Render("j/k or ↑/↓: scroll • esc/backspace: back • q: quit")
}

func getThreadColor(depth int) lipgloss.Color {
	colors := []string{
		"#FF6B6B",
		"#4ECDC4",
		"#45B7D1",
		"#FFA07A",
		"#98D8C8",
		"#F7DC6F",
		"#BB8FCE",
		"#85C1E2",
	}
	return lipgloss.Color(colors[depth%len(colors)])
}

func renderComment(comment reddit.Comment, maxWidth int) string {
	var s strings.Builder

	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))

	indent := ""
	for i := 0; i < comment.Depth; i++ {
		color := getThreadColor(i)
		style := lipgloss.NewStyle().Foreground(color)
		indent += style.Render("│ ")
	}

	s.WriteString(indent)
	s.WriteString(metaStyle.Render(comment.Author))
	s.WriteString(" ")
	s.WriteString(metaStyle.Render(fmt.Sprintf("↑%d", comment.Score)))
	s.WriteString("\n")

	bodyLines := strings.Split(wrapText(comment.Body, maxWidth-(comment.Depth*2)), "\n")
	for _, line := range bodyLines {
		s.WriteString(indent)
		s.WriteString(line)
		s.WriteString("\n")
	}

	if len(comment.Replies) > 0 {
		for _, reply := range comment.Replies {
			s.WriteString(renderComment(reply, maxWidth))
		}
	} else if comment.Depth == 0 {
		s.WriteString("\n")
	}

	return s.String()
}

var feedCmd = &cobra.Command{
	Use:   "feed",
	Short: "List hot posts",
	Run: func(cmd *cobra.Command, args []string) {
		// Load saved theme
		loadSavedTheme(cmd.Context())

		posts := reddit.FetchFeeds(cmd.Context())

		if len(posts) == 0 {
			fmt.Println("\nNo posts found. Subscribe to some subreddits first!")
			return
		}

		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Score > posts[j].Score
		})

		p := tea.NewProgram(model{posts: posts, ctx: cmd.Context()}, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v", err)
		}
	},
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)

		if lineLen+wordLen+1 > width {
			if lineLen > 0 {
				result.WriteString("\n")
				lineLen = 0
			}

			if wordLen > width {
				result.WriteString(word[:width])
				result.WriteString("\n")
				continue
			}
		}

		if lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}

		result.WriteString(word)
		lineLen += wordLen

		if i < len(words)-1 && lineLen+1+len(words[i+1]) > width {
			result.WriteString("\n")
			lineLen = 0
		}
	}

	return result.String()
}

func loadSavedTheme(ctx context.Context) {
	if gormDB := db.FromContext(ctx); gormDB != nil {
		if themeName, err := db.GetSetting(gormDB, "theme"); err == nil && themeName != "" {
			SetTheme(themeName)
		}
	}
}

func init() {
	applyTheme()
	rootCmd.AddCommand(feedCmd)
}
