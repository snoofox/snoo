package cmd

import (
	"html"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func truncate(s string, maxLen int) string {
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func displaySourceName(sourceName string) string {
	if idx := strings.LastIndex(sourceName, ":"); idx != -1 {
		sourceName = sourceName[:idx]
	}
	return sourceName
}

func wrapText(text string, width int) string {
	text = html.UnescapeString(text)

	if width <= 0 {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)

		if lineLen+wordLen+1 > width {
			if lineLen > 0 {
				result.WriteString("\n")
				lineLen = 0
			}

			if wordLen > width {
				result.WriteString(word[:width])
				result.WriteString("\n")
				continue
			}
		}

		if lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}

		result.WriteString(word)
		lineLen += wordLen

		if i < len(words)-1 && lineLen+1+len(words[i+1]) > width {
			result.WriteString("\n")
			lineLen = 0
		}
	}

	return result.String()
}

func renderMarkdown(text string, width int) string {
	if text == "" {
		return ""
	}

	text = html.UnescapeString(text)

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return wrapText(text, width)
	}

	rendered, err := r.Render(text)
	if err != nil {
		return wrapText(text, width)
	}
	result := strings.TrimSpace(rendered)
	return result
}

func getThreadColor(depth int) lipgloss.Color {
	colors := []string{
		"#FF6B6B",
		"#4ECDC4",
		"#45B7D1",
		"#FFA07A",
		"#98D8C8",
		"#F7DC6F",
		"#BB8FCE",
		"#85C1E2",
	}
	return lipgloss.Color(colors[depth%len(colors)])
}
