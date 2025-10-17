package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"snoo/src/db"
	"sync"
	"time"

	"gorm.io/gorm"
)

const BASE_URL = "https://www.reddit.com"

func Purge(ctx context.Context) {
	database := db.FromContext(ctx)
	oneWeekAgo := time.Now().Add(-7 * 24 * time.Hour)

	database.Where("created_at < ?", oneWeekAgo).Delete(&db.Post{})
	database.Where("created_at < ?", oneWeekAgo).Delete(&db.Comment{})
}

func FetchSubreddit(name string) (Subreddit, error) {
	url := fmt.Sprintf("%s/r/%s/about.json", BASE_URL, name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Subreddit{}, fmt.Errorf("error creating request: %s", err)
	}

	req.Header.Set("User-Agent", "snoo:v1.0.0")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return Subreddit{}, fmt.Errorf("error fetching subreddit: %s", err)
	}
	defer resp.Body.Close()

	jsonBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return Subreddit{}, fmt.Errorf("error reading response body: %s", err)
	}

	var raw any
	err = json.Unmarshal(jsonBytes, &raw)

	if err != nil {
		return Subreddit{}, fmt.Errorf("error unmarshalling response body: %s", err)
	}

	var rawSubreddit any = raw.(map[string]any)["data"]
	return Subreddit{
		RedditID:    rawSubreddit.(map[string]any)["name"].(string),
		Name:        rawSubreddit.(map[string]any)["url"].(string),
		DisplayName: rawSubreddit.(map[string]any)["display_name"].(string),
		Desc:        rawSubreddit.(map[string]any)["public_description"].(string),
		Subscribers: rawSubreddit.(map[string]any)["subscribers"].(float64),
		NSFW:        rawSubreddit.(map[string]any)["over18"].(bool),
		IconURL:     rawSubreddit.(map[string]any)["icon_img"].(string),
		CreatedUTC:  rawSubreddit.(map[string]any)["created"].(float64),
		LastFetchAt: nil,
	}, nil
}

func fetchOrGetCachedPosts(database *gorm.DB, subreddit db.Subreddit) []Post {
	now := time.Now()
	needsFetch := subreddit.LastFetchAt == nil || now.Sub(*subreddit.LastFetchAt) > time.Hour

	if !needsFetch {
		var cachedPosts []db.Post
		database.Where("subreddit_name = ?", subreddit.Name).Find(&cachedPosts)

		if len(cachedPosts) > 0 {
			posts := make([]Post, len(cachedPosts))
			for i, p := range cachedPosts {
				posts[i] = Post{
					ID:          p.RedditID,
					Title:       p.Title,
					Author:      p.Author,
					Subreddit:   p.SubredditName,
					Permalink:   p.Permalink,
					URL:         p.URL,
					Score:       p.Score,
					NumComments: p.NumComments,
					CreatedUTC:  p.CreatedUTC,
					IsSelf:      p.IsSelf,
					Selftext:    p.Selftext,
					Thumbnail:   p.Thumbnail,
					NSFW:        p.NSFW,
				}
			}
			return posts
		}
	}

	posts, err := fetchSubredditPosts(subreddit.Name)
	if err != nil {
		fmt.Printf("Error fetching posts from %s: %v\n", subreddit.Name, err)
		return []Post{}
	}

	database.Model(&subreddit).Update("last_fetch_at", now)
	database.Where("subreddit_name = ?", subreddit.Name).Delete(&db.Post{})

	for _, post := range posts {
		dbPost := db.Post{
			RedditID:      post.ID,
			Title:         post.Title,
			Author:        post.Author,
			SubredditID:   subreddit.ID,
			SubredditName: subreddit.Name,
			Permalink:     post.Permalink,
			URL:           post.URL,
			Score:         post.Score,
			NumComments:   post.NumComments,
			CreatedUTC:    post.CreatedUTC,
			IsSelf:        post.IsSelf,
			Selftext:      post.Selftext,
			Thumbnail:     post.Thumbnail,
			NSFW:          post.NSFW,
		}
		database.Create(&dbPost)
	}

	return posts
}

func fetchSubredditPosts(name string) ([]Post, error) {
	url := fmt.Sprintf("%s/%s.json", BASE_URL, name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	req.Header.Set("User-Agent", "snoo:v1.0.0")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error fetching subreddit: %s", err)
	}
	defer resp.Body.Close()

	jsonBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	var raw any
	err = json.Unmarshal(jsonBytes, &raw)

	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %s", err)
	}

	posts := []Post{}
	data := raw.(map[string]any)["data"].(map[string]any)
	children := data["children"].([]any)

	for _, child := range children {
		childData := child.(map[string]any)["data"].(map[string]any)

		post := Post{
			ID:          childData["id"].(string),
			Title:       childData["title"].(string),
			Author:      childData["author"].(string),
			Subreddit:   childData["subreddit"].(string),
			Permalink:   childData["permalink"].(string),
			URL:         childData["url"].(string),
			Score:       int(childData["score"].(float64)),
			NumComments: int(childData["num_comments"].(float64)),
			CreatedUTC:  childData["created_utc"].(float64),
			IsSelf:      childData["is_self"].(bool),
			NSFW:        childData["over_18"].(bool),
		}

		if selftext, ok := childData["selftext"].(string); ok {
			post.Selftext = selftext
		}

		if thumbnail, ok := childData["thumbnail"].(string); ok {
			post.Thumbnail = thumbnail
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func FetchFeeds(ctx context.Context) []Post {
	database := db.FromContext(ctx)

	var subreddits []db.Subreddit
	result := database.Find(&subreddits)

	if result.Error != nil {
		fmt.Printf("Error fetching subreddits: %v\n", result.Error)
		return []Post{}
	}

	if len(subreddits) == 0 {
		fmt.Println("No subscribed subreddits")
		return []Post{}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	allPosts := []Post{}

	for _, sub := range subreddits {
		wg.Add(1)
		go func(subreddit db.Subreddit) {
			defer wg.Done()
			posts := fetchOrGetCachedPosts(database, subreddit)

			mu.Lock()
			allPosts = append(allPosts, posts...)
			mu.Unlock()
		}(sub)
	}

	wg.Wait()
	return allPosts
}

func FetchComments(ctx context.Context, permalink string) ([]Comment, error) {
	database := db.FromContext(ctx)

	var post db.Post
	result := database.Where("permalink = ?", permalink).First(&post)
	if result.Error != nil {
		return fetchCommentsFromAPI(permalink)
	}

	now := time.Now()
	needsFetch := post.CommentsFetchAt == nil || now.Sub(*post.CommentsFetchAt) > 30*time.Minute

	if !needsFetch {
		var cachedComments []db.Comment
		database.Where("post_id = ? AND parent_id IS NULL", post.ID).Order("score DESC").Find(&cachedComments)

		if len(cachedComments) > 0 {
			comments := make([]Comment, len(cachedComments))
			for i, c := range cachedComments {
				comments[i] = buildCommentTree(database, c)
			}
			return comments, nil
		}
	}

	comments, err := fetchCommentsFromAPI(permalink)
	if err != nil {
		return nil, err
	}

	database.Model(&post).Update("comments_fetch_at", now)
	database.Where("post_id = ?", post.ID).Delete(&db.Comment{})

	saveCommentsToCache(database, comments, post.ID, nil)

	return comments, nil
}

func fetchCommentsFromAPI(permalink string) ([]Comment, error) {
	url := fmt.Sprintf("%s%s.json", BASE_URL, permalink)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	req.Header.Set("User-Agent", "snoo:v1.0.0")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error fetching comments: %s", err)
	}
	defer resp.Body.Close()

	jsonBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	var raw []any
	err = json.Unmarshal(jsonBytes, &raw)

	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %s", err)
	}

	if len(raw) < 2 {
		return []Comment{}, nil
	}

	commentsData := raw[1].(map[string]any)["data"].(map[string]any)
	children := commentsData["children"].([]any)

	comments := []Comment{}
	for _, child := range children {
		childMap, ok := child.(map[string]any)
		if !ok {
			continue
		}

		kind, _ := childMap["kind"].(string)
		if kind != "t1" {
			continue
		}

		childData, ok := childMap["data"].(map[string]any)
		if !ok {
			continue
		}

		if childData["author"] == nil || childData["body"] == nil {
			continue
		}

		comment := parseComment(childData, 0)
		if comment.Body != "" && comment.Author != "" {
			comments = append(comments, comment)
		}
	}

	return comments, nil
}

func saveCommentsToCache(database *gorm.DB, comments []Comment, postID uint, parentID *uint) {
	for _, comment := range comments {
		dbComment := db.Comment{
			RedditID:   comment.ID,
			PostID:     postID,
			ParentID:   parentID,
			Author:     comment.Author,
			Body:       comment.Body,
			Score:      comment.Score,
			CreatedUTC: comment.CreatedUTC,
			Depth:      comment.Depth,
		}
		database.Create(&dbComment)

		if len(comment.Replies) > 0 {
			saveCommentsToCache(database, comment.Replies, postID, &dbComment.ID)
		}
	}
}

func buildCommentTree(database *gorm.DB, dbComment db.Comment) Comment {
	comment := Comment{
		ID:         dbComment.RedditID,
		Author:     dbComment.Author,
		Body:       dbComment.Body,
		Score:      dbComment.Score,
		CreatedUTC: dbComment.CreatedUTC,
		Depth:      dbComment.Depth,
		Replies:    []Comment{},
	}

	var replies []db.Comment
	database.Where("parent_id = ?", dbComment.ID).Order("score DESC").Find(&replies)

	for _, reply := range replies {
		comment.Replies = append(comment.Replies, buildCommentTree(database, reply))
	}

	return comment
}

func parseComment(data map[string]any, depth int) Comment {
	id, _ := data["id"].(string)
	author, _ := data["author"].(string)
	body, _ := data["body"].(string)
	score, _ := data["score"].(float64)
	createdUTC, _ := data["created_utc"].(float64)

	comment := Comment{
		ID:         id,
		Author:     author,
		Body:       body,
		Score:      int(score),
		CreatedUTC: createdUTC,
		Depth:      depth,
		Replies:    []Comment{},
	}

	if replies, ok := data["replies"].(map[string]any); ok {
		if repliesData, ok := replies["data"].(map[string]any); ok {
			if children, ok := repliesData["children"].([]any); ok {
				for _, child := range children {
					if childData, ok := child.(map[string]any)["data"].(map[string]any); ok {
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
