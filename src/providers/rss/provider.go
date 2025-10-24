package rss

import (
	"context"
	"fmt"
	"snoo/src/debug"
	"snoo/src/feed"
	"time"

	"github.com/mmcdole/gofeed"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Type() string {
	return "rss"
}

func (p *Provider) FetchPosts(ctx context.Context, source feed.Source) ([]feed.Post, error) {
	debug.Log("RSS: Fetching from %s", source.Identifier)

	fp := gofeed.NewParser()
	rssFeed, err := fp.ParseURL(source.Identifier)
	if err != nil {
		debug.Log("RSS: Error parsing feed: %v", err)
		return nil, fmt.Errorf("error parsing feed: %w", err)
	}

	debug.Log("RSS: Feed has %d items", len(rssFeed.Items))

	posts := make([]feed.Post, 0, len(rssFeed.Items))
	for i, item := range rssFeed.Items {
		author := "Unknown"
		if item.Author != nil && item.Author.Name != "" {
			author = item.Author.Name
		} else if len(item.Authors) > 0 && item.Authors[0].Name != "" {
			author = item.Authors[0].Name
		} else if item.DublinCoreExt != nil && len(item.DublinCoreExt.Creator) > 0 {
			author = item.DublinCoreExt.Creator[0]
		}

		pubDate := time.Now()
		if item.PublishedParsed != nil {
			pubDate = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			pubDate = *item.UpdatedParsed
		}

		content := item.Description
		if item.Content != "" {
			content = item.Content
		}

		guid := item.GUID
		if guid == "" {
			guid = item.Link
		}

		if i < 5 {
			debug.Log("RSS Item %d: GUID=%s, Title=%s, Link=%s", i, guid, item.Title, item.Link)
		}

		posts = append(posts, feed.Post{
			ID:          guid,
			Title:       item.Title,
			Author:      author,
			SourceName:  fmt.Sprintf("rss/%s", source.Name),
			SourceType:  "rss",
			Permalink:   item.Link,
			URL:         item.Link,
			Score:       0,
			NumComments: 0,
			CreatedAt:   pubDate,
			Content:     content,
			Thumbnail:   "",
			NSFW:        false,
		})
	}

	debug.Log("RSS: Returning %d posts", len(posts))
	return posts, nil
}

func (p *Provider) FetchComments(ctx context.Context, post feed.Post) ([]feed.Comment, error) {
	return []feed.Comment{}, nil
}

func (p *Provider) ValidateSource(ctx context.Context, identifier string) (*feed.SourceMetadata, error) {
	fp := gofeed.NewParser()
	rssFeed, err := fp.ParseURL(identifier)
	if err != nil {
		return nil, fmt.Errorf("error parsing feed: %w", err)
	}

	description := rssFeed.Description
	if description == "" && rssFeed.ITunesExt != nil {
		description = rssFeed.ITunesExt.Summary
	}

	iconURL := ""
	if rssFeed.Image != nil {
		iconURL = rssFeed.Image.URL
	}

	return &feed.SourceMetadata{
		Name:        rssFeed.Title,
		DisplayName: rssFeed.Title,
		Description: description,
		IconURL:     iconURL,
	}, nil
}
