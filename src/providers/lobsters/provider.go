package lobsters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"snoo/src/feed"
	"time"
)

const baseURL = "https://lobste.rs"

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Type() string {
	return "lobsters"
}

func (p *Provider) FetchPosts(ctx context.Context, source feed.Source) ([]feed.Post, error) {
	var url string
	if source.Identifier == "active" || source.Identifier == "recent" {
		url = fmt.Sprintf("%s/%s.json", baseURL, source.Identifier)
	} else {
		return nil, fmt.Errorf("invalid lobsters category: %s (use 'active' or 'recent')", source.Identifier)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("User-Agent", "snoo:v1.0.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching lobsters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lobsters returned status %d", resp.StatusCode)
	}

	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var stories []LobstersStory
	if err := json.Unmarshal(jsonBytes, &stories); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	posts := make([]feed.Post, 0, len(stories))
	for _, story := range stories {
		post := parsePost(story, source.Identifier)
		posts = append(posts, post)
	}

	return posts, nil
}

func (p *Provider) FetchComments(ctx context.Context, post feed.Post) ([]feed.Comment, error) {
	url := fmt.Sprintf("%s/s/%s.json", baseURL, post.ID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("User-Agent", "snoo:v1.0.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching comments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lobsters returned status %d", resp.StatusCode)
	}

	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var story LobstersStory
	if err := json.Unmarshal(jsonBytes, &story); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	comments := make([]feed.Comment, 0)
	for _, c := range story.Comments {
		comment := parseComment(c, 0)
		comments = append(comments, comment)
	}

	return comments, nil
}

func (p *Provider) ValidateSource(ctx context.Context, identifier string) (*feed.SourceMetadata, error) {
	if identifier != "active" && identifier != "recent" {
		return nil, fmt.Errorf("invalid lobsters category: %s (use 'active' or 'recent')", identifier)
	}

	displayName := fmt.Sprintf("Lobsters - %s", identifier)
	description := fmt.Sprintf("Lobste.rs %s stories", identifier)

	return &feed.SourceMetadata{
		Name:        identifier,
		DisplayName: displayName,
		Description: description,
		IconURL:     "https://lobste.rs/apple-touch-icon-144.png",
	}, nil
}

type LobstersStory struct {
	ShortID       string            `json:"short_id"`
	Title         string            `json:"title"`
	URL           string            `json:"url"`
	Score         int               `json:"score"`
	CommentCount  int               `json:"comment_count"`
	Description   string            `json:"description"`
	CommentsURL   string            `json:"comments_url"`
	SubmitterUser string            `json:"submitter_user"`
	CreatedAt     string            `json:"created_at"`
	Tags          []string          `json:"tags"`
	Comments      []LobstersComment `json:"comments"`
}

type LobstersComment struct {
	ShortID        string            `json:"short_id"`
	Comment        string            `json:"comment"`
	CommentPlain   string            `json:"comment_plain"`
	Score          int               `json:"score"`
	CreatedAt      string            `json:"created_at"`
	CommentingUser string            `json:"commenting_user"`
	IndentLevel    int               `json:"indent_level"`
	Comments       []LobstersComment `json:"comments"`
}

func parsePost(story LobstersStory, category string) feed.Post {
	createdAt, _ := time.Parse(time.RFC3339, story.CreatedAt)

	return feed.Post{
		ID:          story.ShortID,
		Title:       story.Title,
		Author:      story.SubmitterUser,
		SourceName:  fmt.Sprintf("lobsters/%s", category),
		SourceType:  "lobsters",
		Permalink:   story.CommentsURL,
		URL:         story.URL,
		Score:       story.Score,
		NumComments: story.CommentCount,
		CreatedAt:   createdAt,
		Content:     story.Description,
		Thumbnail:   "",
		NSFW:        false,
	}
}

func parseComment(c LobstersComment, depth int) feed.Comment {
	createdAt, _ := time.Parse(time.RFC3339, c.CreatedAt)

	comment := feed.Comment{
		ID:        c.ShortID,
		Author:    c.CommentingUser,
		Body:      c.CommentPlain,
		Score:     c.Score,
		CreatedAt: createdAt,
		Depth:     depth,
		Replies:   []feed.Comment{},
	}

	for _, reply := range c.Comments {
		comment.Replies = append(comment.Replies, parseComment(reply, depth+1))
	}

	return comment
}
