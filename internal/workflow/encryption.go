package workflow

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

const (
	encryptionEnvKey = "storage_secret"
	encryptionPrefix = "ENCv1:"
)

func encryptionKey() ([]byte, bool) {
	secret := os.Getenv(encryptionEnvKey)
	if secret == "" {
		return nil, false
	}
	sum := sha256.Sum256([]byte(secret))
	return sum[:], true
}

func maybeEncrypt(data []byte) ([]byte, error) {
	key, ok := encryptionKey()
	if !ok {
		return data, nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, data, nil)
	payload := append(nonce, sealed...)
	encoded := base64.StdEncoding.EncodeToString(payload)
	out := append([]byte(encryptionPrefix), []byte(encoded)...)
	return out, nil
}

func maybeDecrypt(data []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(data)
	if !bytes.HasPrefix(trimmed, []byte(encryptionPrefix)) {
		return data, nil
	}
	key, ok := encryptionKey()
	if !ok {
		return nil, errors.New("storage_secret required to read encrypted history")
	}
	payload, err := base64.StdEncoding.DecodeString(string(trimmed[len(encryptionPrefix):]))
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(payload) < nonceSize {
		return nil, errors.New("encrypted history corrupt or truncated")
	}
	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
