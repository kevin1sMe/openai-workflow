package workflow

import (
	"strings"
)

func ResolveChatModel(gptModel, override string) string {
	if override != "" {
		return override
	}
	return gptModel
}

func TrimContext(messages []Message, max int) []Message {
	if max <= 0 || len(messages) <= max {
		return messages
	}
	return messages[len(messages)-max:]
}

func BuildMessages(systemPrompt string, context []Message) []map[string]string {
	var msgs []map[string]string
	if systemPrompt != "" {
		msgs = append(msgs, map[string]string{"role": "system", "content": systemPrompt})
	}
	for _, m := range context {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	return msgs
}

func JoinStrings(parts ...string) string {
	return strings.Join(parts, "")
}
