package template

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

// Fetcher defines how to get template content.
type Fetcher interface {
	ReadFile(path string) ([]byte, error)
	Walk(root string, fn fs.WalkDirFunc) error
}

// EmbeddedFetcher uses the compiled-in binary data.
type EmbeddedFetcher struct {
	FS fs.FS
}

func (e *EmbeddedFetcher) ReadFile(path string) ([]byte, error) {
	return fs.ReadFile(e.FS, path)
}

func (e *EmbeddedFetcher) Walk(root string, fn fs.WalkDirFunc) error {
	return fs.WalkDir(e.FS, root, fn)
}

// SmartFetcher implements "Local-Override, Embed-Default" strategy.
type SmartFetcher struct {
	Embedded *EmbeddedFetcher
	LocalDir string
	Logger   *slog.Logger
}

func NewSmartFetcher(embeddedFS fs.FS, logger *slog.Logger) *SmartFetcher {
	home, _ := os.UserHomeDir()
	return &SmartFetcher{
		Embedded: &EmbeddedFetcher{FS: embeddedFS},
		LocalDir: filepath.Join(home, ".helix", "templates"),
		Logger:   logger,
	}
}

func (s *SmartFetcher) ReadFile(path string) ([]byte, error) {
	// Try Local Override
	// Only check if file exists locally. No network calls.
	localPath := filepath.Join(s.LocalDir, path)
	if info, err := os.Stat(localPath); err == nil && !info.IsDir() {
		s.Logger.Debug("Using local template override", "path", localPath)
		return os.ReadFile(localPath)
	}

	// Fallback to Embedded (Default/Safe)
	return s.Embedded.ReadFile(path)
}

func (s *SmartFetcher) Walk(root string, fn fs.WalkDirFunc) error {
	// Only walk embedded. Overriding entire directory structures is too complex for now.
	return s.Embedded.Walk(root, fn)
}

func NetworkCheck() {
	// No-op
}
