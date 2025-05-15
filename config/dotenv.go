package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// FileReader is an interface for reading files
type FileReader interface {
	Open(name string) (io.ReadCloser, error)
}

// OSFileReader implements FileReader using os package
type OSFileReader struct{}

// Open opens a file using os.Open
func (r *OSFileReader) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// DefaultFileReader is the default file reader
var DefaultFileReader FileReader = &OSFileReader{}

// LoadEnvFile loads environment variables from a .env file if it exists
func LoadEnvFile() {
	LoadEnvFileWithReader(DefaultFileReader, viper.GetViper())
}

// LoadEnvFileWithReader loads environment variables from a .env file using the provided reader
func LoadEnvFileWithReader(reader FileReader, v *viper.Viper) {
	envFile, err := reader.Open(".env")
	if err != nil {
		cwd, _ := os.Getwd()
		envFile, err = os.Open(filepath.Join(cwd, ".env"))
		if err != nil {
			return
		}
	}

	defer func() {
		if err := envFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error closing .env file: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"'`)

		if strings.HasPrefix(key, "N8N_") {
			viperKey := strings.ToLower(strings.TrimPrefix(key, "N8N_"))
			if os.Getenv(key) == "" {
				v.Set(viperKey, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error reading .env file: %v\n", err)
	}
}
