package db

import (
	"time"

	"gorm.io/gorm"
)

// Source represents a unified feed source (subreddit, RSS feed, etc.)
type Source struct {
	gorm.Model
	Type        string `gorm:"size:32;index"`
	Identifier  string `gorm:"size:512;index"` // subreddit name, RSS URL, etc.
	Name        string `gorm:"size:256"`
	DisplayName string `gorm:"size:256"`
	Description string `gorm:"type:text"`
	IconURL     string `gorm:"size:512"`
	LastFetchAt *time.Time
}

type Post struct {
	gorm.Model
	SourceID        uint `gorm:"index"`
	Source          Source
	SourceType      string `gorm:"size:32;index"`
	ExternalID      string `gorm:"size:512;index"` // Reddit ID, RSS GUID, etc.
	Title           string `gorm:"type:text"`
	Author          string `gorm:"size:128"`
	SourceName      string `gorm:"size:256;index"`
	Permalink       string `gorm:"size:512"`
	URL             string `gorm:"type:text"`
	Score           int
	NumComments     int
	CreatedUTC      float64 `gorm:"index"`
	Content         string  `gorm:"type:text"`
	Thumbnail       string  `gorm:"size:512"`
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

type Setting struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;size:64"`
	Value string `gorm:"size:256"`
}
