package config

import (
	"bufio"
	"os"
	"strings"
)

// loadDotenv reads KEY=VALUE pairs from a .env file in the working
// directory, if one exists, and applies them via os.Setenv — but only for
// keys not already set in the real environment, so `PORT=9000 go run .`
// still wins over whatever .env says. No external dependency: this backend
// otherwise has none (see go.mod), so a ~20-line reader is preferable to
// pulling in godotenv for one file format.
func loadDotenv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, alreadySet := os.LookupEnv(key); !alreadySet {
			os.Setenv(key, value)
		}
	}
}
