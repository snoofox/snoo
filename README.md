# snoo

A terminal-based Reddit reader built with Go

> ⚠️ **Disclaimer**: This is a vibe-coded project. Don't blame me if it sucks.

## Features

- 📱 Browse hot posts from your subscribed subreddits
- 💬 Read comments with threaded replies
- 🎨 Colorful TUI with syntax highlighting
- 📦 No authentication needed

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

## Usage

Run the app (feed is the default command):

```bash
snoo
# or explicitly
snoo feed
```

### Navigation

**In feed list:**
- `j/k` or `↑/↓` - Navigate posts
- `Enter` or `Space` - Open post and comments
- `q` - Quit

**In post view:**
- `j/k` or `↑/↓` - Scroll through comments
- `Esc` or `Backspace` - Back to feed
- `q` - Quit

## Commands

### Subscribe to subreddits

```bash
snoo sub add <subreddit-name>
```

### View your feed

```bash
snoo feed
```

## How it works

- Uses Reddit's public JSON API (no authentication required)
- Stores subscriptions and cached data in a local SQLite database (`data.sqlite3`)
- Posts are cached for 1 hour, comments for 30 minutes
- Old data is automatically cleaned up every 6 hours

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [GORM](https://gorm.io/) - ORM for SQLite
- SQLite - Local database

## License

Do whatever you want with it.
