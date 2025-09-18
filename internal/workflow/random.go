package workflow

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"
)

func RandomUID() string {
    buf := make([]byte, 4)
    if _, err := rand.Read(buf); err == nil {
        return hex.EncodeToString(buf)
    }
    return fmt.Sprintf("%x", time.Now().UnixNano())
}
