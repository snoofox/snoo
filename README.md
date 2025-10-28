# snoo

<p align="center">
  <img src="https://raw.githubusercontent.com/snoofox/snoo/main/assets/space-gopher.png" alt="snoo gopher" style="max-width: 100%; height: auto;"/>
</p>

A terminal-based universal feed reader built with Go

> ⚠️ **Disclaimer**: This is a vibe-coded project. Don't blame me if it sucks.

## Preview
![snoo preview](https://raw.githubusercontent.com/snoofox/snoo/main/assets/demo.gif)

## Features

- 🌐 Multi-provider support - Reddit, RSS feeds, Lobsters, and more coming soon
- 📰 Browse posts from all your subscribed sources in one unified feed
- 🔍 Filter posts by source with interactive toggle UI
- 💬 Read comments with threaded replies and colored thread indicators (Reddit & Lobsters)
- 📖 Read full articles directly in the terminal
- 🎯 Subscribe to different Reddit sorts (hot, new, rising, top, best)
- 🎨 Multiple color themes (Catppuccin, Dracula, GitHub, Peppermint, and more)
- 📦 No authentication needed - uses public APIs
- 💾 Smart caching with SQLite
- ⚡ Fast and lightweight

## Installation

```bash
CGO_ENABLED=1 go install github.com/snoofox/snoo@latest
```

Or build from source:

```bash
git clone <repo-url>
cd snoo
go build
```

## Quick Start

1. Subscribe to some feeds:
```bash
# Add Reddit subreddits (defaults to 'best' sort)
snoo sub add golang
snoo sub add programming:hot
snoo sub add rust:new

# Add RSS feeds
snoo sub rss https://example.com/feed.xml

# Add Lobsters
snoo sub lobsters active
snoo sub lobsters recent
```

2. View your unified feed:
```bash
snoo
# or explicitly
snoo feed
```

3. Change the theme (optional):
```bash
snoo theme catppuccin
```

## Commands

### Managing Subscriptions

Subscribe to a feed:
```bash
# Reddit subreddit (with optional sort)
snoo sub add <subreddit-name>[:sort]
# Examples:
snoo sub add golang          # defaults to 'best'
snoo sub add golang:hot      # hot posts
snoo sub add golang:new      # new posts
snoo sub add golang:rising   # rising posts
snoo sub add golang:top      # top posts

# RSS feed
snoo sub rss <feed-url>

# Lobsters (active or recent)
snoo sub lobsters <category>
```

List your subscriptions:
```bash
snoo sub list
```

Unsubscribe from a feed:
```bash
snoo sub rm <subscription-id> # you can get id from sub list
```

### Viewing Your Feed

View posts from all subscribed sources:
```bash
snoo feed
# or just
snoo
```

### Themes

List available themes:
```bash
snoo theme
```

Change theme:
```bash
snoo theme <theme-name>
```

Available themes:
- `default` - Original pink/purple theme
- `catppuccin` - Catppuccin Mocha palette
- `dracula` - Dracula color scheme
- `github` - GitHub-inspired colors
- `peppermint` - Fresh mint and cyan tones

Your theme preference is saved and persists across sessions.

## Navigation

### In feed list:
- `j` or `↓` - Move down
- `k` or `↑` - Move up
- `Enter` or `Space` - Open post and load comments
- `f` - Open filter menu to toggle sources
- `q` - Quit

### In filter menu:
- `j` or `↓` - Move down
- `k` or `↑` - Move up
- `Space` or `Enter` - Toggle selected source
- `a` - Enable all sources
- `d` - Disable all sources
- `Esc` or `Backspace` - Back to feed

### In post view:
- `j` or `↓` - Scroll down
- `k` or `↑` - Scroll up
- `r` - Read full article (for link posts)
- `Esc` or `Backspace` - Back to feed
- `q` - Quit

## Supported Providers

- **Reddit** - Browse subreddits, read threaded comments
- **RSS** - Subscribe to any RSS/Atom feed
- **Lobsters** - Browse active or recent posts from lobste.rs
- **More coming soon** - 4chan and other platforms planned

## How it works

- Pluggable provider architecture for easy extensibility
- Uses public APIs (no authentication required)
- Stores subscriptions and cached data in a local SQLite database (`data.sqlite3`)
- Posts are cached for 1 hour to reduce API calls
- Duplicate posts across different sources are automatically deduplicated
- Filter sources on-the-fly without losing your subscriptions
- Theme preferences are persisted in the database
- Article content is extracted and rendered as markdown

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [GORM](https://gorm.io/) - ORM for SQLite
- SQLite - Local database

## License

Do whatever you want with it.
