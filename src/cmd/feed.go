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

type postKey struct {
	typ string
	id  string
}

type sortOption struct {
	name string
	key  string
}

var sortOptions = []sortOption{
	{name: "Most Upvotes", key: "upvotes_desc"},
	{name: "Least Upvotes", key: "upvotes_asc"},
	{name: "Newest First", key: "newest"},
	{name: "Oldest First", key: "oldest"},
	{name: "Most Comments", key: "comments_desc"},
	{name: "Least Comments", key: "comments_asc"},
}

type commentSortOption struct {
	name string
	key  string
}

var commentSortOptions = []commentSortOption{
	{name: "Best (Highest Score)", key: "best"},
	{name: "Worst (Lowest Score)", key: "worst"},
	{name: "Newest First", key: "newest"},
	{name: "Oldest First", key: "oldest"},
}

type model struct {
	posts              []Post
	allPosts           []Post
	cursor             int
	viewing            bool
	filtering          bool
	sorting            bool
	commentSorting     bool
	filterCursor       int
	sortCursor         int
	commentSortCursor  int
	selected           int
	width              int
	height             int
	comments           []Comment
	loadingComments    bool
	loadingArticle     bool
	ctx                context.Context
	viewport           viewport.Model
	ready              bool
	sources            []string
	sourceEnabled      map[string]bool
	currentSort        string
	currentCommentSort string
	originalContent    string
	articleContent     string
	showingArticle     bool
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
		m.applyCommentSorting()
		m.loadingComments = false
		m.viewport.SetContent(m.renderPostContent())
		m.viewport.GotoTop()

	case articleLoadedMsg:
		if msg.err != nil {
			debug.Log("Failed to load article: %v", msg.err)
			m.articleContent = fmt.Sprintf("Failed to load article: %v", msg.err)
		} else if msg.content != "" {
			m.articleContent = msg.content
		} else {
			m.articleContent = "No content could be extracted from this article."
		}
		m.showingArticle = true
		m.loadingArticle = false
		m.viewport.SetContent(m.renderPostContent())
		m.viewport.GotoTop()

	case tea.KeyMsg:
		if m.commentSorting {
			switch msg.String() {
			case "q", "esc", "backspace":
				m.commentSorting = false
				m.commentSortCursor = 0
				return m, nil
			case "up", "k":
				if m.commentSortCursor > 0 {
					m.commentSortCursor--
				}
			case "down", "j":
				if m.commentSortCursor < len(commentSortOptions)-1 {
					m.commentSortCursor++
				}
			case "enter", " ":
				m.currentCommentSort = commentSortOptions[m.commentSortCursor].key
				m.applyCommentSorting()
				m.commentSorting = false
				m.commentSortCursor = 0
				m.viewport.SetContent(m.renderPostContent())
				m.viewport.GotoTop()
				m.savePreferences()
				return m, nil
			}
		} else if m.sorting {
			switch msg.String() {
			case "q", "esc", "backspace":
				m.sorting = false
				m.sortCursor = 0
				return m, nil
			case "up", "k":
				if m.sortCursor > 0 {
					m.sortCursor--
				}
			case "down", "j":
				if m.sortCursor < len(sortOptions)-1 {
					m.sortCursor++
				}
			case "enter", " ":
				m.currentSort = sortOptions[m.sortCursor].key
				m.applySorting()
				m.sorting = false
				m.sortCursor = 0
				m.savePreferences()
				return m, nil
			}
		} else if m.filtering {
			switch msg.String() {
			case "q", "esc", "backspace":
				m.filtering = false
				m.filterCursor = 0
				return m, nil
			case "up", "k":
				if m.filterCursor > 0 {
					m.filterCursor--
				}
			case "down", "j":
				if m.filterCursor < len(m.sources)-1 {
					m.filterCursor++
				}
			case "enter", " ":
				source := m.sources[m.filterCursor]
				m.sourceEnabled[source] = !m.sourceEnabled[source]
				m.applyFilters()
				m.savePreferences()
				return m, nil
			case "a":
				for _, source := range m.sources {
					m.sourceEnabled[source] = true
				}
				m.applyFilters()
				m.savePreferences()
				return m, nil
			case "d":
				for _, source := range m.sources {
					m.sourceEnabled[source] = false
				}
				m.applyFilters()
				m.savePreferences()
				return m, nil
			}
		} else if m.viewing {
			switch msg.String() {
			case "q", "esc", "backspace":
				m.viewing = false
				m.comments = nil
				m.loadingComments = false
				m.loadingArticle = false
				m.originalContent = ""
				m.articleContent = ""
				m.showingArticle = false
				return m, nil
			case "s":
				if len(m.comments) > 0 {
					m.commentSorting = true
					m.commentSortCursor = 0
					return m, nil
				}
			case "r":
				if !m.loadingArticle {
					post := m.posts[m.selected]
					if post.URL != "" {
						// If we already have article content, toggle between original and article
						if m.articleContent != "" {
							m.showingArticle = !m.showingArticle
							m.viewport.SetContent(m.renderPostContent())
							m.viewport.GotoTop()
						} else {
							// First time fetching article, save original content
							m.originalContent = post.Content
							m.loadingArticle = true
							m.viewport.SetContent(m.renderPostContent())
							return m, m.loadArticleCmd()
						}
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
			case "f":
				m.filtering = true
				m.filterCursor = 0
				return m, nil
			case "s":
				m.sorting = true
				m.sortCursor = 0
				return m, nil
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
				m.originalContent = ""
				m.articleContent = ""
				m.showingArticle = false
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

func (m *model) applyFilters() {
	m.posts = m.posts[:0]
	for i := range m.allPosts {
		if m.sourceEnabled[m.allPosts[i].SourceName] {
			m.posts = append(m.posts, m.allPosts[i])
		}
	}
	m.applySorting()
	if m.cursor >= len(m.posts) {
		m.cursor = len(m.posts) - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
}

func (m *model) applySorting() {
	switch m.currentSort {
	case "upvotes_desc":
		sort.Slice(m.posts, func(i, j int) bool {
			return m.posts[i].Score > m.posts[j].Score
		})
	case "upvotes_asc":
		sort.Slice(m.posts, func(i, j int) bool {
			return m.posts[i].Score < m.posts[j].Score
		})
	case "newest":
		sort.Slice(m.posts, func(i, j int) bool {
			return m.posts[i].CreatedUTC > m.posts[j].CreatedUTC
		})
	case "oldest":
		sort.Slice(m.posts, func(i, j int) bool {
			return m.posts[i].CreatedUTC < m.posts[j].CreatedUTC
		})
	case "comments_desc":
		sort.Slice(m.posts, func(i, j int) bool {
			return m.posts[i].NumComments > m.posts[j].NumComments
		})
	case "comments_asc":
		sort.Slice(m.posts, func(i, j int) bool {
			return m.posts[i].NumComments < m.posts[j].NumComments
		})
	default:
		// Default: smart sort (upvotes if available, otherwise newest)
		sort.Slice(m.posts, func(i, j int) bool {
			if m.posts[i].Score > 0 && m.posts[j].Score > 0 {
				return m.posts[i].Score > m.posts[j].Score
			}
			return m.posts[i].CreatedUTC > m.posts[j].CreatedUTC
		})
	}
}

func (m *model) applyCommentSorting() {
	m.comments = sortComments(m.comments, m.currentCommentSort)
}

func sortComments(comments []Comment, sortKey string) []Comment {
	if len(comments) == 0 {
		return comments
	}

	sorted := make([]Comment, len(comments))
	copy(sorted, comments)

	switch sortKey {
	case "best":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Score > sorted[j].Score
		})
	case "worst":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Score < sorted[j].Score
		})
	case "newest":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt > sorted[j].CreatedAt
		})
	case "oldest":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt < sorted[j].CreatedAt
		})
	}

	for i := range sorted {
		if len(sorted[i].Replies) > 0 {
			sorted[i].Replies = sortComments(sorted[i].Replies, sortKey)
		}
	}

	return sorted
}

func (m *model) savePreferences() {
	database := db.FromContext(m.ctx)
	if database == nil {
		return
	}

	if m.currentSort != "" {
		db.SetSetting(database, "feed_sort", m.currentSort)
	}

	if m.currentCommentSort != "" {
		db.SetSetting(database, "comment_sort", m.currentCommentSort)
	}

	var enabledSources []string
	for _, src := range m.sources {
		if m.sourceEnabled[src] {
			enabledSources = append(enabledSources, src)
		}
	}
	if len(enabledSources) > 0 {
		db.SetSetting(database, "feed_sources", strings.Join(enabledSources, ","))
	}
}

func loadPreferences(ctx context.Context, sources []string) (string, string, map[string]bool) {
	database := db.FromContext(ctx)
	sourceEnabled := make(map[string]bool, len(sources))

	for _, src := range sources {
		sourceEnabled[src] = true
	}

	if database == nil {
		return "upvotes_desc", "best", sourceEnabled
	}

	sortPref, err := db.GetSetting(database, "feed_sort")
	if err != nil || sortPref == "" {
		sortPref = "upvotes_desc"
	}

	commentSortPref, err := db.GetSetting(database, "comment_sort")
	if err != nil || commentSortPref == "" {
		commentSortPref = "best"
	}

	sourcesStr, err := db.GetSetting(database, "feed_sources")
	if err == nil && sourcesStr != "" {
		enabledList := strings.Split(sourcesStr, ",")
		for _, src := range sources {
			sourceEnabled[src] = false
		}
		for _, enabled := range enabledList {
			if _, exists := sourceEnabled[enabled]; exists {
				sourceEnabled[enabled] = true
			}
		}
	}

	return sortPref, commentSortPref, sourceEnabled
}

func (m model) View() string {
	if m.commentSorting {
		return m.viewCommentSort()
	}
	if m.sorting {
		return m.viewSort()
	}
	if m.filtering {
		return m.viewFilter()
	}
	if m.viewing {
		return m.viewPost()
	}
	return m.viewList()
}

func (m model) viewCommentSort() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Sort Comments"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Choose comment sorting order"))
	b.WriteString("\n\n")

	for i, opt := range commentSortOptions {
		indicator := " "
		if opt.key == m.currentCommentSort {
			indicator = "●"
		}

		line := fmt.Sprintf("  %s %s", indicator, opt.name)
		if i == m.commentSortCursor {
			b.WriteString(cursorStyle.Render("● "))
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString("  ")
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	t := GetCurrentTheme()
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpNav).Render("j/k"))
	b.WriteString(dimStyle.Render(" navigate  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpAction).Render("enter"))
	b.WriteString(dimStyle.Render(" select  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpQuit).Render("esc"))
	b.WriteString(dimStyle.Render(" back"))

	return b.String()
}

func (m model) viewSort() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Sort Posts"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Choose sorting order"))
	b.WriteString("\n\n")

	for i, opt := range sortOptions {
		indicator := " "
		if opt.key == m.currentSort {
			indicator = "●"
		}

		line := fmt.Sprintf("  %s %s", indicator, opt.name)
		if i == m.sortCursor {
			b.WriteString(cursorStyle.Render("● "))
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString("  ")
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	t := GetCurrentTheme()
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpNav).Render("j/k"))
	b.WriteString(dimStyle.Render(" navigate  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpAction).Render("enter"))
	b.WriteString(dimStyle.Render(" select  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpQuit).Render("esc"))
	b.WriteString(dimStyle.Render(" back"))

	return b.String()
}

func (m model) viewFilter() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Filter Sources"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Toggle sources on/off"))
	b.WriteString("\n\n")

	for i, src := range m.sources {
		box := "○"
		if m.sourceEnabled[src] {
			box = "●"
		}

		line := fmt.Sprintf("  %s %s", box, src)
		if i == m.filterCursor {
			b.WriteString(cursorStyle.Render("● "))
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString("  ")
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	t := GetCurrentTheme()
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpNav).Render("j/k"))
	b.WriteString(dimStyle.Render(" navigate  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpAction).Render("space"))
	b.WriteString(dimStyle.Render(" toggle  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpAction).Render("a"))
	b.WriteString(dimStyle.Render(" all  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpAction).Render("d"))
	b.WriteString(dimStyle.Render(" none  "))
	b.WriteString(lipgloss.NewStyle().Foreground(t.HelpQuit).Render("esc"))
	b.WriteString(dimStyle.Render(" back"))

	return b.String()
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
		lipgloss.NewStyle().Foreground(theme.HelpAction).Render("f") +
		dimStyle.Render(" filter  ") +
		lipgloss.NewStyle().Foreground(theme.HelpAction).Render("s") +
		dimStyle.Render(" sort  ") +
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
	} else if m.showingArticle && m.articleContent != "" {
		// Show fetched article content
		rendered := renderMarkdown(m.articleContent, maxWidth)
		s += rendered + "\n"
	} else if post.Content != "" || m.originalContent != "" {
		// Show original content (either from post or saved original)
		content := post.Content
		if m.originalContent != "" {
			content = m.originalContent
		}
		rendered := renderMarkdown(content, maxWidth)
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
	if post.URL != "" && !m.loadingArticle {
		if m.articleContent != "" {
			if m.showingArticle {
				helpText += " • r: show original"
			} else {
				helpText += " • r: show article"
			}
		} else {
			helpText += " • r: read article"
		}
	}
	if len(m.comments) > 0 {
		helpText += " • s: sort comments"
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

func renderComment(c Comment, w int) string {
	var b strings.Builder

	meta := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))

	var indent strings.Builder
	for i := 0; i < c.Depth; i++ {
		indent.WriteString(lipgloss.NewStyle().Foreground(getThreadColor(i)).Render("│ "))
	}
	ind := indent.String()

	b.WriteString(ind)
	b.WriteString(meta.Render(c.Author))
	b.WriteString(" ")
	b.WriteString(meta.Render(fmt.Sprintf("↑%d", c.Score)))
	b.WriteString("\n")

	body := renderMarkdown(c.Body, w-(c.Depth*2))
	lines := strings.SplitSeq(body, "\n")
	for line := range lines {
		b.WriteString(ind)
		b.WriteString(line)
		b.WriteString("\n")
	}

	if len(c.Replies) > 0 {
		for i := range c.Replies {
			b.WriteString(renderComment(c.Replies[i], w))
		}
	} else if c.Depth == 0 {
		b.WriteString("\n")
	}

	return b.String()
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

		seen := make(map[postKey]bool, len(posts))
		deduped := make([]Post, 0, len(posts))
		for i := range posts {
			k := postKey{posts[i].SourceType, posts[i].ID}
			if !seen[k] {
				seen[k] = true
				deduped = append(deduped, posts[i])
			}
		}
		posts = deduped

		srcSet := make(map[string]bool, 10)
		for i := range posts {
			srcSet[posts[i].SourceName] = true
		}
		srcs := make([]string, 0, len(srcSet))
		for src := range srcSet {
			srcs = append(srcs, src)
		}
		sort.Strings(srcs)

		sortPref, commentSortPref, srcEnabled := loadPreferences(cmd.Context(), srcs)

		m := model{
			posts:              posts,
			allPosts:           posts,
			ctx:                cmd.Context(),
			sources:            srcs,
			sourceEnabled:      srcEnabled,
			currentSort:        sortPref,
			currentCommentSort: commentSortPref,
		}

		m.applyFilters()

		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v", err)
		}
	},
}

func truncate(s string, maxLen int) string {
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func wrapText(text string, width int) string {
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
