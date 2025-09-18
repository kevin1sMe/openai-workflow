package workflow

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type StreamState struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason,omitempty"`
	Error        string `json:"error,omitempty"`
}

func WriteStreamState(path string, state StreamState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	payload, err := maybeEncrypt(data)
	if err != nil {
		return err
	}
	return atomicWrite(path, payload)
}

func ReadStreamState(path string) (StreamState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return StreamState{}, err
	}
	if len(data) == 0 {
		return StreamState{}, nil
	}
	decoded, err := maybeDecrypt(data)
	if err != nil {
		return StreamState{}, err
	}
	if len(decoded) == 0 {
		return StreamState{}, nil
	}
	var state StreamState
	if err := json.Unmarshal(decoded, &state); err != nil {
		return StreamState{}, err
	}
	return state, nil
}

func StreamFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func RemoveFiles(paths ...string) {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			continue
		}
	}
}

func Touch(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	return f.Close()
}

func FileAge(path string, now time.Time) (time.Duration, error) {
	mod, err := FileModified(path)
	if err != nil {
		return 0, err
	}
	return now.Sub(mod), nil
}
