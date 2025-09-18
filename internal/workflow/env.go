package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Env struct {
	WorkflowDataDir   string
	WorkflowCacheDir  string
	APIKey            string
	OrgID             string
	ChatAPIEndpoint   string
	DalleAPIEndpoint  string
	GPTModel          string
	ChatModelOverride string
	SystemPrompt      string
	MaxContext        int
	TimeoutSeconds    int
	StreamFile        string
	PIDFile           string
	ChatFile          string
}

func LoadEnv() (*Env, error) {
	dataDir := os.Getenv("alfred_workflow_data")
	cacheDir := os.Getenv("alfred_workflow_cache")
	if dataDir == "" || cacheDir == "" {
		return nil, fmt.Errorf("workflow data/cache dirs not set")
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}

	maxContext := readIntEnv("max_context", 4)
	timeout := readIntEnv("timeout_seconds", 20)

	env := &Env{
		WorkflowDataDir:   dataDir,
		WorkflowCacheDir:  cacheDir,
		APIKey:            os.Getenv("openai_api_key"),
		OrgID:             os.Getenv("openai_org_id"),
		ChatAPIEndpoint:   os.Getenv("chatgpt_api_endpoint"),
		DalleAPIEndpoint:  os.Getenv("dalle_api_endpoint"),
		GPTModel:          os.Getenv("gpt_model"),
		ChatModelOverride: os.Getenv("chatgpt_model_override"),
		SystemPrompt:      os.Getenv("system_prompt"),
		MaxContext:        maxContext,
		TimeoutSeconds:    timeout,
	}
	env.StreamFile = filepath.Join(cacheDir, "stream.txt")
	env.PIDFile = filepath.Join(cacheDir, "pid.txt")
	env.ChatFile = filepath.Join(dataDir, "chat.json")
	return env, nil
}

func readIntEnv(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return i
}
