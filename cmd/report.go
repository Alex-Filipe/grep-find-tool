package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// reportsDir returns where reports are saved: ~/Downloads when it exists,
// falling back to the home directory, then the current directory.
func reportsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	downloads := filepath.Join(home, "Downloads")
	if fi, err := os.Stat(downloads); err == nil && fi.IsDir() {
		return downloads
	}
	return home
}

// saveReport writes content to a timestamped .txt file under reportsDir and
// returns the full path.
func saveReport(prefix, content string) (string, error) {
	name := fmt.Sprintf("grep-tool-%s-%s.txt", prefix, time.Now().Format("2006-01-02-150405"))
	path := filepath.Join(reportsDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}
