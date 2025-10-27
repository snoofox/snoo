package hackernews

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/snoofox/snoo/src/feed"
)

const (
	baseURL = "https://hacker-news.firebaseio.com/v0"
	hnURL   = "https://news.ycombinator.com"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
}

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Type() string {
	return "hackernews"
}

type hnItem struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Time        int64  `json:"time"`
	Text        string `json:"text"`
	Dead        bool   `json:"dead"`
	Deleted     bool   `json:"deleted"`
	Parent      int    `json:"parent"`
	Kids        []int  `json:"kids"`
	URL         string `json:"url"`
	Score       int    `json:"score"`
	Title       string `json:"title"`
	Descendants int    `json:"descendants"`
}

func (p *Provider) FetchPosts(ctx context.Context, source feed.Source) ([]feed.Post, error) {
	var storyIDs []int
	var err error

	switch source.Identifier {
	case "top":
		storyIDs, err = p.fetchStoryIDs("topstories")
	case "new":
		storyIDs, err = p.fetchStoryIDs("newstories")
	case "best":
		storyIDs, err = p.fetchStoryIDs("beststories")
	case "ask":
		storyIDs, err = p.fetchStoryIDs("askstories")
	case "show":
		storyIDs, err = p.fetchStoryIDs("showstories")
	case "job":
		storyIDs, err = p.fetchStoryIDs("jobstories")
	default:
		return nil, fmt.Errorf("unknown HackerNews category: %s", source.Identifier)
	}

	if err != nil {
		return nil, err
	}

	limit := 20
	if len(storyIDs) > limit {
		storyIDs = storyIDs[:limit]
	}

	posts := p.fetchItemsConcurrently(storyIDs, source.Identifier)
	return posts, nil
}

func (p *Provider) FetchComments(ctx context.Context, post feed.Post) ([]feed.Comment, error) {
	var storyID int
	fmt.Sscanf(post.ID, "%d", &storyID)

	item, err := p.fetchItem(storyID)
	if err != nil {
		return nil, err
	}

	if item == nil || len(item.Kids) == 0 {
		return []feed.Comment{}, nil
	}

	maxTopComments := 20
	topKids := item.Kids
	if len(topKids) > maxTopComments {
		topKids = topKids[:maxTopComments]
	}

	comments := p.fetchCommentsConcurrently(topKids, 0)
	return comments, nil
}

func (p *Provider) ValidateSource(ctx context.Context, identifier string) (*feed.SourceMetadata, error) {
	validCategories := map[string]string{
		"top":  "Top Stories",
		"new":  "New Stories",
		"best": "Best Stories",
		"ask":  "Ask HN",
		"show": "Show HN",
		"job":  "Jobs",
	}

	displayName, ok := validCategories[identifier]
	if !ok {
		return nil, fmt.Errorf("invalid category. Valid options: top, new, best, ask, show, job")
	}

	return &feed.SourceMetadata{
		Name:        identifier,
		DisplayName: fmt.Sprintf("HackerNews - %s", displayName),
		Description: fmt.Sprintf("HackerNews %s feed", displayName),
		IconURL:     "https://news.ycombinator.com/favicon.ico",
	}, nil
}

func (p *Provider) fetchStoryIDs(endpoint string) ([]int, error) {
	url := fmt.Sprintf("%s/%s.json", baseURL, endpoint)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching story IDs: %w", err)
	}
	defer resp.Body.Close()

	var ids []int
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, fmt.Errorf("error decoding story IDs: %w", err)
	}

	return ids, nil
}

func (p *Provider) fetchItem(id int) (*hnItem, error) {
	url := fmt.Sprintf("%s/item/%d.json", baseURL, id)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching item: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var item hnItem
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("error decoding item: %w", err)
	}

	return &item, nil
}

func (p *Provider) fetchItemsConcurrently(ids []int, category string) []feed.Post {
	type result struct {
		post feed.Post
		err  error
	}

	results := make(chan result, len(ids))
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, 20)

	for _, id := range ids {
		wg.Add(1)
		go func(itemID int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			item, err := p.fetchItem(itemID)
			if err != nil || item == nil || item.Deleted || item.Dead {
				results <- result{err: err}
				return
			}

			post := p.itemToPost(item, category)
			results <- result{post: post}
		}(id)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	posts := make([]feed.Post, 0, len(ids))
	for res := range results {
		if res.err == nil && res.post.ID != "" {
			posts = append(posts, res.post)
		}
	}

	return posts
}

func (p *Provider) fetchCommentsConcurrently(ids []int, depth int) []feed.Comment {
	type result struct {
		comment *feed.Comment
		index   int
	}

	results := make(chan result, len(ids))
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, 30)

	for i, id := range ids {
		wg.Add(1)
		go func(itemID, idx int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			comment, err := p.fetchCommentTree(itemID, depth)
			if err == nil && comment != nil {
				results <- result{comment: comment, index: idx}
			}
		}(id, i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	commentMap := make(map[int]*feed.Comment)
	for res := range results {
		commentMap[res.index] = res.comment
	}

	comments := make([]feed.Comment, 0, len(ids))
	for i := 0; i < len(ids); i++ {
		if comment, ok := commentMap[i]; ok {
			comments = append(comments, *comment)
		}
	}

	return comments
}

func (p *Provider) fetchCommentTree(id int, depth int) (*feed.Comment, error) {
	item, err := p.fetchItem(id)
	if err != nil || item == nil || item.Deleted || item.Dead {
		return nil, err
	}

	if item.Type != "comment" {
		return nil, nil
	}

	comment := &feed.Comment{
		ID:        fmt.Sprintf("%d", item.ID),
		Author:    item.By,
		Body:      item.Text,
		Score:     item.Score,
		CreatedAt: time.Unix(item.Time, 0),
		Depth:     depth,
		Replies:   []feed.Comment{},
	}

	if depth < 1 && len(item.Kids) > 0 {
		maxReplies := 5
		kids := item.Kids
		if len(kids) > maxReplies {
			kids = kids[:maxReplies]
		}
		comment.Replies = p.fetchCommentsConcurrently(kids, depth+1)
	}

	return comment, nil
}

func (p *Provider) itemToPost(item *hnItem, category string) feed.Post {
	url := item.URL
	if url == "" {
		url = fmt.Sprintf("%s/item?id=%d", hnURL, item.ID)
	}

	return feed.Post{
		ID:          fmt.Sprintf("%d", item.ID),
		Title:       item.Title,
		Author:      item.By,
		SourceName:  fmt.Sprintf("HackerNews/%s", category),
		SourceType:  "hackernews",
		Permalink:   fmt.Sprintf("/item?id=%d", item.ID),
		URL:         url,
		Score:       item.Score,
		NumComments: item.Descendants,
		CreatedAt:   time.Unix(item.Time, 0),
		Content:     item.Text,
	}
}
