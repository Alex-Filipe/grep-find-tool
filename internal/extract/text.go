package extract

import (
	"io"
	"os"
)

func init() {
	exts := []string{".txt", ".md", ".log", ".go", ".json", ".yaml", ".toml", ".css", ".html", ".js", ".ts", ".py", ".rs", ".java", ".c", ".h", ".sh", ".yml"}
	for _, e := range exts {
		Register(e, &textExtractor{})
	}
}

type textExtractor struct{}

func (t *textExtractor) Extract(path string) (io.ReadCloser, error) {
	if isBinary(path) {
		return nil, nil
	}
	return os.Open(path)
}
