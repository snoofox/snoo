package cmd

// Post represents a post for UI display
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
	CreatedUTC  float64
	Content     string
	Thumbnail   string
	NSFW        bool
}

// Comment represents a comment for UI display
type Comment struct {
	ID        string
	Author    string
	Body      string
	Score     int
	CreatedAt float64
	Depth     int
	Replies   []Comment
}
