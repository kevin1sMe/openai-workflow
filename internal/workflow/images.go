package workflow

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var promptLine = regexp.MustCompile(`^(.* Prompt:)`)

func MarkdownImage(path string) string {
	prompt := ExtractPrompt(path)
	if prompt == "" {
		prompt = "**No prompt found**"
	}
	return fmt.Sprintf("%s\n![](%s)", prompt, path)
}

func ExtractPrompt(path string) string {
	text, err := ReadMetadata("kMDItemDescription", path)
	if err != nil || text == "" {
		return ""
	}
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if promptLine.MatchString(line) {
			out = append(out, promptLine.ReplaceAllString(line, "**$1**"))
		}
	}
	return strings.Join(out, "\n\n")
}

func BuildImageFilename(baseDir string, creation time.Time, uid string) string {
	parts := []string{
		fmt.Sprintf("%04d", creation.Year()),
		fmt.Sprintf("%02d", creation.Month()),
		fmt.Sprintf("%02d", creation.Day()),
		fmt.Sprintf("%02d", creation.Hour()),
		fmt.Sprintf("%02d", creation.Minute()),
		fmt.Sprintf("%02d", creation.Second()),
	}
	name := strings.Join(parts, ".") + "-" + uid + ".png"
	return filepath.Join(baseDir, name)
}
