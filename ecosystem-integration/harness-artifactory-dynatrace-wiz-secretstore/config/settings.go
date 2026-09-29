package config

import (
	"os"
	"regexp"
	"strings"
)

var (
	TODO_DBFILE = getEnv("TODO_DBFILE", "scheduler.db")
	TODO_PORT   = getEnv("TODO_PORT", "7540")
)

// GetPassword returns the application password. It prefers the TODO_PASSWORD
// environment variable and falls back to the Vault-injected file at
// /vault/secrets/config.txt when the env var is not set.
func GetPassword() string {
	if v := os.Getenv("TODO_PASSWORD"); v != "" {
		return v
	}
	b, err := os.ReadFile("/vault/secrets/config.txt")
	if err == nil {
		s := strings.TrimSpace(string(b))
		if s != "" {
			// Try to extract `todoPassword` from formats like:
			// data: map[todoPassword:test12345]
			re := regexp.MustCompile(`todoPassword\s*[:=]\s*([^\]\s\n]+)`)
			if m := re.FindStringSubmatch(s); len(m) >= 2 {
				return m[1]
			}
			// Fallback: if file is just the password
			return s
		}
	}
	return ""
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
