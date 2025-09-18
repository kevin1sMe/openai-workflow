package workflow

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func EnsureChatFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return WriteChat(path, []Message{})
}

func ReadChat(path string) ([]Message, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Message{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []Message{}, nil
	}
	decoded, err := maybeDecrypt(data)
	if err != nil {
		return nil, err
	}
	if len(decoded) == 0 {
		return []Message{}, nil
	}
	var messages []Message
	if err := json.Unmarshal(decoded, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func WriteChat(path string, msgs []Message) error {
	data, err := json.Marshal(msgs)
	if err != nil {
		return err
	}
	payload, err := maybeEncrypt(data)
	if err != nil {
		return err
	}
	return atomicWrite(path, payload)
}

func AppendChat(path string, msg Message) error {
	msgs, err := ReadChat(path)
	if err != nil {
		return err
	}
	msgs = append(msgs, msg)
	return WriteChat(path, msgs)
}

func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func FileModified(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
