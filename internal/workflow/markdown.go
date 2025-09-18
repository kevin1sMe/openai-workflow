package workflow

import "strings"

func MarkdownChat(messages []Message, ignoreLastInterrupted bool) string {
	var builder strings.Builder
	for i, msg := range messages {
		switch msg.Role {
		case "assistant":
			if msg.Content != "" {
				builder.WriteString(msg.Content)
				builder.WriteString("\n\n")
			}
		case "user":
			builder.WriteString("# ⊙ You\n\n")
			builder.WriteString(msg.Content)
			builder.WriteString("\n\n# ⊚ Assistant\n\n")
			userTwice := i+1 < len(messages) && messages[i+1].Role == "user"
			last := i == len(messages)-1
			if userTwice || (last && !ignoreLastInterrupted) {
				builder.WriteString("[Answer Interrupted]\n\n")
			}
		}
	}
	return strings.TrimSpace(builder.String())
}
