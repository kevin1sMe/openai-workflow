package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DalleEnv struct {
	ParentFolder    string
	Model           string
	Style           string
	Quality         string
	IncludeMetadata bool
}

func LoadDalleEnv() (*DalleEnv, error) {
	folder := os.Getenv("dalle_images_folder")
	if folder == "" {
		return nil, fmt.Errorf("dalle_images_folder not set")
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return nil, err
	}
	include := stringsEqualFold(os.Getenv("dalle_write_metadata"), "1", "true", "yes")
	return &DalleEnv{
		ParentFolder:    folder,
		Model:           os.Getenv("dalle_model"),
		Style:           os.Getenv("dalle_style"),
		Quality:         os.Getenv("dalle_quality"),
		IncludeMetadata: include,
	}, nil
}

func stringsEqualFold(value string, accepted ...string) bool {
	for _, v := range accepted {
		if strings.EqualFold(value, v) {
			return true
		}
	}
	return false
}

func LatestImages(folder string, max int) ([]string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			files = append(files, filepath.Join(folder, entry.Name()))
		}
	}
	sort.Strings(files)
	if len(files) > max {
		files = files[len(files)-max:]
	}
	return files, nil
}
