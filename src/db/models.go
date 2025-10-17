package db

import (
	"time"

	"gorm.io/gorm"
)

type Subreddit struct {
	gorm.Model
	RedditID    string `gorm:"uniqueIndex;size:12"`
	Name        string `gorm:"uniqueIndex;size:64"`
	DisplayName string `gorm:"size:128"`
	Desc        string `gorm:"type:text"`
	Subscribers float64
	NSFW        bool
	IconURL     string
	CreatedUTC  float64
	LastFetchAt *time.Time
}

type Post struct {
	gorm.Model
	RedditID        string `gorm:"uniqueIndex;size:12"`
	Title           string `gorm:"type:text"`
	Author          string `gorm:"size:64"`
	SubredditID     uint
	Subreddit       Subreddit
	SubredditName   string `gorm:"size:64;index"`
	Permalink       string `gorm:"size:512"`
	URL             string `gorm:"type:text"`
	Score           int
	NumComments     int
	CreatedUTC      float64
	IsSelf          bool
	Selftext        string `gorm:"type:text"`
	Thumbnail       string `gorm:"size:512"`
	NSFW            bool
	CommentsFetchAt *time.Time
}

type Comment struct {
	gorm.Model
	RedditID   string `gorm:"uniqueIndex;size:12"`
	PostID     uint   `gorm:"index"`
	Post       Post
	ParentID   *uint
	Author     string `gorm:"size:64"`
	Body       string `gorm:"type:text"`
	Score      int
	CreatedUTC float64
	Depth      int
}
