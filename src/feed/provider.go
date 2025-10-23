package feed

import (
	"context"
	"time"
)

type Provider interface {
	Type() string
	FetchPosts(ctx context.Context, source Source) ([]Post, error)
	FetchComments(ctx context.Context, post Post) ([]Comment, error)
	ValidateSource(ctx context.Context, identifier string) (*SourceMetadata, error)
}

type Source struct {
	ID          uint
	Type        string
	Identifier  string // subreddit name, RSS URL, etc.
	Name        string
	DisplayName string
	Description string
	IconURL     string
	Metadata    map[string]interface{}
	LastFetchAt *time.Time
}

type SourceMetadata struct {
	Name        string
	DisplayName string
	Description string
	IconURL     string
	Metadata    map[string]interface{}
}

type Post struct {
	ID          string
	Title       string
	Author      string
	SourceName  string
	SourceType  string
	Permalink   string
	URL         string
	Score       int
	NumComments int
	CreatedAt   time.Time
	Content     string
	Thumbnail   string
	NSFW        bool
}

type Comment struct {
	ID        string
	Author    string
	Body      string
	Score     int
	CreatedAt time.Time
	Depth     int
	Replies   []Comment
}
