package reddit

import "time"

type Subreddit struct {
	RedditID    string `gorm:"uniqueIndex;size:12"`
	Name        string `gorm:"uniqueIndex;size:64"`
	DisplayName string `gorm:"size:128"`
	Desc        string `gorm:"type:text"` // public description
	Subscribers float64
	NSFW        bool
	IconURL     string
	CreatedUTC  float64
	LastFetchAt *time.Time
}

type Post struct {
	ID          string
	Title       string
	Author      string
	Subreddit   string
	Permalink   string
	URL         string
	Score       int
	NumComments int
	CreatedUTC  float64
	IsSelf      bool
	Selftext    string
	Thumbnail   string
	NSFW        bool
}

type Comment struct {
	ID         string
	Author     string
	Body       string
	Score      int
	CreatedUTC float64
	Depth      int
	Replies    []Comment
}
