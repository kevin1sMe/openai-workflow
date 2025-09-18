package workflow

import (
    "net/url"
    "strings"
)

func NormalizeBaseURL(raw, defaultBase string, trimSuffixes ...string) string {
    if raw == "" {
        return defaultBase
    }
    parsed, err := url.Parse(raw)
    if err != nil || parsed.Scheme == "" {
        return defaultBase
    }
    for _, suffix := range trimSuffixes {
        if parsed.Path == suffix {
            parsed.Path = ""
            break
        }
        if strings.HasSuffix(parsed.Path, suffix) {
            parsed.Path = strings.TrimSuffix(parsed.Path, suffix)
            break
        }
    }
    if parsed.Path == "" || parsed.Path == "/" {
        parsed.Path = "/v1"
    }
    return strings.TrimRight(parsed.String(), "/")
}
