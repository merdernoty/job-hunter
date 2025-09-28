package env

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open env file %s: %w", filename, err)
	}
	defer file.Close()

	var loaded int
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Printf("Warning: invalid format at line %d in %s: %s", lineNum, filename, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"'`)

		if os.Getenv(key) == "" {
			os.Setenv(key, value)
			loaded++
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file %s: %w", filename, err)
	}

	log.Printf("Loaded %d environment variables from %s", loaded, filename)
	return nil
}

func Load() error {
	return LoadFromFile(".env")
}

func MustLoad() {
	if err := Load(); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
}
func LoadIfExists() {
	if err := Load(); err != nil {
		log.Printf("Warning: %v", err)
	}
}