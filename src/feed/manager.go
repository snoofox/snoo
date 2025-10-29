package feed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/snoofox/snoo/src/db"
	"github.com/snoofox/snoo/src/debug"
	"gorm.io/gorm"
)

type Manager struct {
	db *gorm.DB
}

func NewManager(database *gorm.DB) *Manager {
	return &Manager{db: database}
}

func (m *Manager) FetchAll(ctx context.Context) ([]Post, error) {
	var sources []db.Source
	if err := m.db.Find(&sources).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sources: %w", err)
	}

	if len(sources) == 0 {
		return []Post{}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	allPosts := []Post{}

	for _, src := range sources {
		wg.Add(1)
		go func(source db.Source) {
			defer wg.Done()

			provider, err := Get(source.Type)
			if err != nil {
				fmt.Printf("Unknown provider type %s: %v\n", source.Type, err)
				return
			}

			feedSource := dbSourceToFeedSource(source)
			posts, err := m.fetchOrGetCached(ctx, provider, feedSource)
			if err != nil {
				fmt.Printf("Error fetching from %s: %v\n", source.Name, err)
				return
			}

			mu.Lock()
			allPosts = append(allPosts, posts...)
			mu.Unlock()
		}(src)
	}

	wg.Wait()
	return allPosts, nil
}

func (m *Manager) fetchOrGetCached(ctx context.Context, provider Provider, source Source) ([]Post, error) {
	now := time.Now()
	needsFetch := source.LastFetchAt == nil || now.Sub(*source.LastFetchAt) > time.Hour

	if !needsFetch {
		var cachedPosts []db.Post
		m.db.Where("source_id = ?", source.ID).Order("created_utc DESC").Find(&cachedPosts)

		if len(cachedPosts) > 0 {
			posts := make([]Post, len(cachedPosts))
			for i, p := range cachedPosts {
				posts[i] = dbPostToFeedPost(p)
			}
			return posts, nil
		}
	}

	posts, err := provider.FetchPosts(ctx, source)
	if err != nil {
		debug.Log("Error fetching posts from %s: %v", source.Name, err)
		return nil, err
	}

	debug.Log("Fetched %d posts from %s (%s)", len(posts), source.Name, source.Type)

	m.db.Model(&db.Source{}).Where("id = ?", source.ID).Update("last_fetch_at", now)

	savedCount := 0
	for i, post := range posts {
		if post.ID == "" {
			debug.Log("Post %d has empty ID, title: %s", i, post.Title)
			continue
		}

		dbPost := feedPostToDBPost(post, source.ID)

		var existing db.Post
		result := m.db.Where("external_id = ? AND source_id = ?", post.ID, source.ID).First(&existing)
		if result.Error == nil {
			m.db.Model(&existing).Updates(dbPost)
		} else {
			if err := m.db.Create(&dbPost).Error; err != nil {
				debug.Log("Error creating post: %v", err)
			} else {
				savedCount++
			}
		}
	}

	debug.Log("Saved %d new posts from %s", savedCount, source.Name)

	return posts, nil
}

func (m *Manager) Subscribe(ctx context.Context, providerType, identifier string) error {
	provider, err := Get(providerType)
	if err != nil {
		return err
	}

	var existing db.Source
	if err := m.db.Where("type = ? AND identifier = ?", providerType, identifier).First(&existing).Error; err == nil {
		return fmt.Errorf("already subscribed to this source")
	}

	metadata, err := provider.ValidateSource(ctx, identifier)
	if err != nil {
		return fmt.Errorf("failed to validate source: %w", err)
	}

	source := &db.Source{
		Type:        providerType,
		Identifier:  metadata.Name,
		Name:        metadata.Name,
		DisplayName: metadata.DisplayName,
		Description: metadata.Description,
		IconURL:     metadata.IconURL,
	}

	if err := m.db.Create(source).Error; err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	return nil
}

func (m *Manager) Unsubscribe(id uint) error {
	result := m.db.Unscoped().Delete(&db.Source{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("source not found")
	}
	return nil
}

func (m *Manager) ListSources() ([]Source, error) {
	var dbSources []db.Source
	if err := m.db.Find(&dbSources).Error; err != nil {
		return nil, err
	}

	sources := make([]Source, len(dbSources))
	for i, s := range dbSources {
		sources[i] = dbSourceToFeedSource(s)
	}
	return sources, nil
}

func (m *Manager) FetchComments(ctx context.Context, post Post) ([]Comment, error) {
	provider, err := Get(post.SourceType)
	if err != nil {
		return nil, err
	}

	return provider.FetchComments(ctx, post)
}

func dbSourceToFeedSource(s db.Source) Source {
	return Source{
		ID:          s.ID,
		Type:        s.Type,
		Identifier:  s.Identifier,
		Name:        s.Name,
		DisplayName: s.DisplayName,
		Description: s.Description,
		IconURL:     s.IconURL,
		LastFetchAt: s.LastFetchAt,
	}
}

func dbPostToFeedPost(p db.Post) Post {
	return Post{
		ID:          p.ExternalID,
		Title:       p.Title,
		Author:      p.Author,
		SourceName:  p.SourceName,
		SourceType:  p.SourceType,
		Permalink:   p.Permalink,
		URL:         p.URL,
		Score:       p.Score,
		NumComments: p.NumComments,
		CreatedAt:   time.Unix(int64(p.CreatedUTC), 0),
		Content:     p.Content,
		Thumbnail:   p.Thumbnail,
		NSFW:        p.NSFW,
	}
}

func feedPostToDBPost(p Post, sourceID uint) db.Post {
	return db.Post{
		SourceID:    sourceID,
		SourceType:  p.SourceType,
		ExternalID:  p.ID,
		Title:       p.Title,
		Author:      p.Author,
		SourceName:  p.SourceName,
		Permalink:   p.Permalink,
		URL:         p.URL,
		Score:       p.Score,
		NumComments: p.NumComments,
		CreatedUTC:  float64(p.CreatedAt.Unix()),
		Content:     p.Content,
		Thumbnail:   p.Thumbnail,
		NSFW:        p.NSFW,
	}
}
