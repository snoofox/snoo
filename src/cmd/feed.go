package cmd

import (
	"context"
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/snoofox/snoo/src/article"
	"github.com/snoofox/snoo/src/db"
	"github.com/snoofox/snoo/src/debug"
	"github.com/snoofox/snoo/src/feed"
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
	comments []Comment
}

type articleLoadedMsg struct {
	content string
	err     error
}

type model struct {
	posts           []Post
	cursor          int
	viewing         bool
	selected        int
	width           int
	height          int
	comments        []Comment
	loadingComments bool
	loadingArticle  bool
	ctx             context.Context
	viewport        viewport.Model
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
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 2
		}

	case commentsLoadedMsg:
		m.comments = msg.comments
		m.loadingComments = false
		m.viewport.SetContent(m.renderPostContent())
		m.viewport.GotoTop()

	case articleLoadedMsg:
		if msg.err != nil {
			debug.Log("Failed to load article: %v", msg.err)
			m.posts[m.selected].Content = fmt.Sprintf("Failed to load article: %v", msg.err)
		} else if msg.content != "" {
			m.posts[m.selected].Content = msg.content
		} else {
			m.posts[m.selected].Content = "No content could be extracted from this article."
		}
		m.loadingArticle = false
		m.viewport.SetContent(m.renderPostContent())
		m.viewport.GotoTop()

	case tea.KeyMsg:
		if m.viewing {
			switch msg.String() {
			case "q", "esc", "backspace":
				m.viewing = false
				m.comments = nil
				m.loadingComments = false
				m.loadingArticle = false
				return m, nil
			case "r":
				if !m.loadingArticle {
					post := m.posts[m.selected]
					if post.URL != "" && post.Content == "" {
						m.loadingArticle = true
						m.viewport.SetContent(m.renderPostContent())
						return m, m.loadArticleCmd()
					}
				}
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
				}
			case "down", "j":
				if m.cursor < len(m.posts)-1 {
					m.cursor++
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

func (m model) viewList() string {
	if !m.ready {
		return "Loading..."
	}

	headerLines := 4
	linesPerPost := 3
	availableLines := m.height - 2

	firstVisiblePost := 0
	if m.cursor > 0 {
		middlePost := (availableLines / linesPerPost) / 3
		firstVisiblePost = m.cursor - middlePost
		if firstVisiblePost < 0 {
			firstVisiblePost = 0
		}
	}

	maxVisiblePosts := (availableLines - headerLines) / linesPerPost
	lastVisiblePost := firstVisiblePost + maxVisiblePosts
	if lastVisiblePost > len(m.posts) {
		lastVisiblePost = len(m.posts)
	}

	s := "\n"
	s += titleStyle.Render("  󰑍  Your Feed") + "\n"
	s += dimStyle.Render(fmt.Sprintf("  %d posts", len(m.posts))) + "\n\n"

	for i := firstVisiblePost; i < lastVisiblePost; i++ {
		post := m.posts[i]
		titleText := truncate(post.Title, 85)

		if m.cursor == i {
			cursor := cursorStyle.Render("● ")
			nsfw := ""
			if post.NSFW {
				nsfw = nsfwStyle.Render(" NSFW ") + " "
			}
			sub := subredditStyle.Render(post.SourceName)

			metadata := ""
			sep := separatorStyle.Render("•")

			if post.Score > 0 {
				metadata += scoreStyle.Render(fmt.Sprintf(" %d", post.Score))
			}

			if post.NumComments > 0 {
				if metadata != "" {
					metadata += " " + sep + " "
				}
				metadata += commentsStyle.Render(fmt.Sprintf("󰆉 %d", post.NumComments))
			}

			s += " " + cursor + selectedStyle.Render(titleText) + "\n"
			if metadata != "" {
				s += "   " + nsfw + sub + " " + sep + " " + metadata + "\n\n"
			} else {
				s += "   " + nsfw + sub + "\n\n"
			}
		} else {
			nsfw := ""
			if post.NSFW {
				nsfw = nsfwStyle.Render(" NSFW ") + " "
			}
			sub := subredditStyle.Render(post.SourceName)

			metadata := ""
			sep := separatorStyle.Render("•")

			if post.Score > 0 {
				metadata += scoreStyle.Render(fmt.Sprintf(" %d", post.Score))
			}

			if post.NumComments > 0 {
				if metadata != "" {
					metadata += " " + sep + " "
				}
				metadata += commentsStyle.Render(fmt.Sprintf("󰆉 %d", post.NumComments))
			}

			s += "   " + lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB")).Render(titleText) + "\n"
			if metadata != "" {
				s += "   " + nsfw + sub + " " + sep + " " + metadata + "\n\n"
			} else {
				s += "   " + nsfw + sub + "\n\n"
			}
		}
	}

	theme := GetCurrentTheme()
	helpText := dimStyle.Render("  ") +
		lipgloss.NewStyle().Foreground(theme.HelpNav).Render("j/k") +
		dimStyle.Render(" navigate  ") +
		lipgloss.NewStyle().Foreground(theme.HelpAction).Render("enter") +
		dimStyle.Render(" read  ") +
		lipgloss.NewStyle().Foreground(theme.HelpQuit).Render("q") +
		dimStyle.Render(" quit")
	return s + helpText
}

func (m model) loadCommentsCmd() tea.Cmd {
	return func() tea.Msg {
		post := m.posts[m.selected]
		if post.SourceType == "reddit" || post.SourceType == "lobsters" || post.SourceType == "hackernews" {
			database := db.FromContext(m.ctx)
			manager := feed.NewManager(database)

			feedPost := feed.Post{
				ID:         post.ID,
				SourceType: post.SourceType,
				Permalink:  post.Permalink,
			}

			feedComments, err := manager.FetchComments(m.ctx, feedPost)
			if err != nil {
				return commentsLoadedMsg{comments: []Comment{}}
			}

			comments := make([]Comment, len(feedComments))
			for i, c := range feedComments {
				comments[i] = convertComment(c)
			}
			return commentsLoadedMsg{comments: comments}
		}
		return commentsLoadedMsg{comments: []Comment{}}
	}
}

func (m model) loadArticleCmd() tea.Cmd {
	return func() tea.Msg {
		post := m.posts[m.selected]
		if post.URL != "" {
			content, err := article.Fetch(m.ctx, post.URL)
			return articleLoadedMsg{content: content, err: err}
		}
		return articleLoadedMsg{content: "", err: fmt.Errorf("no URL available")}
	}
}

func convertComment(c feed.Comment) Comment {
	replies := make([]Comment, len(c.Replies))
	for i, r := range c.Replies {
		replies[i] = convertComment(r)
	}
	return Comment{
		ID:        c.ID,
		Author:    c.Author,
		Body:      c.Body,
		Score:     c.Score,
		CreatedAt: float64(c.CreatedAt.Unix()),
		Depth:     c.Depth,
		Replies:   replies,
	}
}

func (m model) renderPostContent() string {
	post := m.posts[m.selected]
	maxWidth := m.width - 4
	if maxWidth < 40 {
		maxWidth = 40
	}

	s := "\n" + titleStyle.Render(wrapText(post.Title, maxWidth)) + "\n\n"

	if m.loadingArticle {
		s += dimStyle.Render("Loading article...") + "\n"
	} else if post.Content != "" {
		rendered := renderMarkdown(post.Content, maxWidth)
		s += rendered + "\n"
	} else {
		s += urlStyle.Render(post.URL) + "\n"
		if post.URL != "" {
			s += "\n" + dimStyle.Render("Press 'r' to read full article") + "\n"
		}
	}

	if post.SourceType == "reddit" || post.SourceType == "lobsters" || post.SourceType == "hackernews" {
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
	}

	return s
}

func (m model) viewPost() string {
	post := m.posts[m.selected]
	helpText := "j/k or ↑/↓: scroll"
	if post.URL != "" && post.Content == "" && !m.loadingArticle {
		helpText += " • r: read article"
	}
	helpText += " • esc/backspace: back • q: quit"
	return m.viewport.View() + "\n" + dimStyle.Render(helpText)
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

func renderComment(comment Comment, maxWidth int) string {
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

	rendered := renderMarkdown(comment.Body, maxWidth-(comment.Depth*2))
	bodyLines := strings.SplitSeq(rendered, "\n")
	for line := range bodyLines {
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
	Short: "List hot posts from all sources",
	Run: func(cmd *cobra.Command, args []string) {
		loadSavedTheme(cmd.Context())

		database := db.FromContext(cmd.Context())
		manager := feed.NewManager(database)

		feedPosts, err := manager.FetchAll(cmd.Context())
		if err != nil {
			fmt.Printf("Error fetching feeds: %v\n", err)
			return
		}

		if len(feedPosts) == 0 {
			fmt.Println("\nNo posts found. Subscribe to some sources first!")
			fmt.Println("Try: snoo sub add golang")
			fmt.Println("     snoo sub rss https://example.com/feed.xml")
			return
		}

		posts := make([]Post, len(feedPosts))
		for i, p := range feedPosts {
			posts[i] = Post{
				ID:          p.ID,
				Title:       p.Title,
				Author:      p.Author,
				SourceName:  p.SourceName,
				SourceType:  p.SourceType,
				Permalink:   p.Permalink,
				URL:         p.URL,
				Score:       p.Score,
				NumComments: p.NumComments,
				CreatedUTC:  float64(p.CreatedAt.Unix()),
				Content:     p.Content,
				Thumbnail:   p.Thumbnail,
				NSFW:        p.NSFW,
			}
		}

		sort.Slice(posts, func(i, j int) bool {
			if posts[i].Score > 0 && posts[j].Score > 0 {
				return posts[i].Score > posts[j].Score
			}
			return posts[i].CreatedUTC > posts[j].CreatedUTC
		})

		p := tea.NewProgram(model{posts: posts, ctx: cmd.Context()}, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v", err)
		}
	},
}

func truncate(s string, maxLen int) string {
	// Decode HTML entities first
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func wrapText(text string, width int) string {
	// Decode HTML entities first
	text = html.UnescapeString(text)

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

func renderMarkdown(text string, width int) string {
	if text == "" {
		return ""
	}

	// Decode HTML entities like &amp; -> &, &lt; -> <, etc.
	text = html.UnescapeString(text)

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return wrapText(text, width)
	}

	rendered, err := r.Render(text)
	if err != nil {
		return wrapText(text, width)
	}
	result := strings.TrimSpace(rendered)
	return result
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
