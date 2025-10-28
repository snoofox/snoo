package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/snoofox/snoo/src/feed"
)

const baseURL = "https://www.reddit.com"

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Type() string {
	return "reddit"
}

func (p *Provider) FetchPosts(ctx context.Context, source feed.Source) ([]feed.Post, error) {
	subreddit, sort := parseIdentifier(source.Identifier)
	url := fmt.Sprintf("%s/r/%s/%s.json", baseURL, subreddit, sort)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("User-Agent", "snoo:v1.0.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching subreddit: %w", err)
	}
	defer resp.Body.Close()

	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	data, ok := raw["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	children, ok := data["children"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	posts := make([]feed.Post, 0, len(children))
	for _, child := range children {
		childMap, ok := child.(map[string]interface{})
		if !ok {
			continue
		}

		childData, ok := childMap["data"].(map[string]interface{})
		if !ok {
			continue
		}

		post := parsePost(childData, source.Identifier)
		if post.ID != "" {
			posts = append(posts, post)
		}
	}

	return posts, nil
}

func (p *Provider) FetchComments(ctx context.Context, post feed.Post) ([]feed.Comment, error) {
	url := fmt.Sprintf("%s%s.json", baseURL, post.Permalink)

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

	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var raw []interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	if len(raw) < 2 {
		return []feed.Comment{}, nil
	}

	commentsData, ok := raw[1].(map[string]interface{})
	if !ok {
		return []feed.Comment{}, nil
	}

	data, ok := commentsData["data"].(map[string]interface{})
	if !ok {
		return []feed.Comment{}, nil
	}

	children, ok := data["children"].([]interface{})
	if !ok {
		return []feed.Comment{}, nil
	}

	comments := make([]feed.Comment, 0)
	for _, child := range children {
		childMap, ok := child.(map[string]interface{})
		if !ok {
			continue
		}

		kind, _ := childMap["kind"].(string)
		if kind != "t1" {
			continue
		}

		childData, ok := childMap["data"].(map[string]interface{})
		if !ok {
			continue
		}

		comment := parseComment(childData, 0)
		if comment.Body != "" {
			comments = append(comments, comment)
		}
	}

	return comments, nil
}

func (p *Provider) ValidateSource(ctx context.Context, identifier string) (*feed.SourceMetadata, error) {
	subreddit, sort := parseIdentifier(identifier)

	validSorts := map[string]bool{
		"hot": true, "new": true, "rising": true, "top": true, "best": true,
	}
	if !validSorts[sort] {
		return nil, fmt.Errorf("invalid sort type: %s (use hot, new, rising, top, or best)", sort)
	}

	url := fmt.Sprintf("%s/r/%s/about.json", baseURL, subreddit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("User-Agent", "snoo:v1.0.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching subreddit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subreddit not found or unavailable")
	}

	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	data, ok := raw["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	displayName, _ := data["display_name"].(string)
	desc, _ := data["public_description"].(string)
	iconURL, _ := data["icon_img"].(string)

	normalizedIdentifier := fmt.Sprintf("%s:%s", subreddit, sort)

	return &feed.SourceMetadata{
		Name:        normalizedIdentifier,
		DisplayName: fmt.Sprintf("r/%s (%s)", displayName, sort),
		Description: desc,
		IconURL:     iconURL,
	}, nil
}

func parseIdentifier(identifier string) (subreddit, sort string) {
	subreddit = identifier
	sort = "best"

	for i := 0; i < len(identifier); i++ {
		if identifier[i] == ':' {
			subreddit = identifier[:i]
			sort = identifier[i+1:]
			break
		}
	}

	return subreddit, sort
}

func parsePost(data map[string]interface{}, subreddit string) feed.Post {
	id, _ := data["id"].(string)
	title, _ := data["title"].(string)
	author, _ := data["author"].(string)
	permalink, _ := data["permalink"].(string)
	url, _ := data["url"].(string)
	score, _ := data["score"].(float64)
	numComments, _ := data["num_comments"].(float64)
	createdUTC, _ := data["created_utc"].(float64)
	isSelf, _ := data["is_self"].(bool)
	selftext, _ := data["selftext"].(string)
	thumbnail, _ := data["thumbnail"].(string)
	nsfw, _ := data["over_18"].(bool)

	content := ""
	if isSelf {
		content = selftext
	}

	return feed.Post{
		ID:          id,
		Title:       title,
		Author:      author,
		SourceName:  fmt.Sprintf("r/%s", subreddit),
		SourceType:  "reddit",
		Permalink:   permalink,
		URL:         url,
		Score:       int(score),
		NumComments: int(numComments),
		CreatedAt:   time.Unix(int64(createdUTC), 0),
		Content:     content,
		Thumbnail:   thumbnail,
		NSFW:        nsfw,
	}
}

func parseComment(data map[string]interface{}, depth int) feed.Comment {
	id, _ := data["id"].(string)
	author, _ := data["author"].(string)
	body, _ := data["body"].(string)
	score, _ := data["score"].(float64)
	createdUTC, _ := data["created_utc"].(float64)

	comment := feed.Comment{
		ID:        id,
		Author:    author,
		Body:      body,
		Score:     int(score),
		CreatedAt: time.Unix(int64(createdUTC), 0),
		Depth:     depth,
		Replies:   []feed.Comment{},
	}

	if replies, ok := data["replies"].(map[string]interface{}); ok {
		if repliesData, ok := replies["data"].(map[string]interface{}); ok {
			if children, ok := repliesData["children"].([]interface{}); ok {
				for _, child := range children {
					if childData, ok := child.(map[string]interface{})["data"].(map[string]interface{}); ok {
						if childData["author"] != nil && childData["body"] != nil {
							reply := parseComment(childData, depth+1)
							comment.Replies = append(comment.Replies, reply)
						}
					}
				}
			}
		}
	}

	return comment
}
