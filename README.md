# snoo

A terminal-based Reddit reader built with Go

![snoo preview](https://i.ibb.co/tp0SGpb6/snoo.png)

> ⚠️ **Disclaimer**: This is a vibe-coded project. Don't blame me if it sucks.

## Features

- � Browrse hot posts from your subscribed subreddits
- � Read tcomments with threaded replies and colored thread indicators
- 🎨 Multiple color themes (Catppuccin, Dracula, GitHub, Peppermint, and more)
- 📦 No authentication needed - uses Reddit's public JSON API
- 💾 Smart caching with SQLite
- ⚡ Fast and lightweight

## Installation

```bash
go install
```

Or build from source:

```bash
git clone <repo-url>
cd snoo
go build
```

## Quick Start

1. Subscribe to some subreddits:
```bash
snoo sub add golang
snoo sub add programming
snoo sub add linux
```

2. View your feed:
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

Subscribe to a subreddit:
```bash
snoo sub add <subreddit-name>
```

List your subscriptions:
```bash
snoo sub list
```

Unsubscribe from a subreddit:
```bash
snoo sub remove <subreddit-id> # you can get id from sub list
```

### Viewing Your Feed

View posts from subscribed subreddits:
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
- `q` - Quit

### In post view:
- `j` or `↓` - Scroll down
- `k` or `↑` - Scroll up
- `Esc` or `Backspace` - Back to feed
- `q` - Quit

## How it works

- Uses Reddit's public JSON API (no authentication required)
- Stores subscriptions and cached data in a local SQLite database (`data.sqlite3`)
- Posts are cached for 1 hour, comments for 30 minutes
- Old data is automatically cleaned up every 6 hours
- Theme preferences are persisted in the database

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [GORM](https://gorm.io/) - ORM for SQLite
- SQLite - Local database

## License

Do whatever you want with it.
