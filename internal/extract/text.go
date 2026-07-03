package extract

import (
	"bufio"
	"bytes"
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
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(f)
	head, err := br.Peek(512)
	if err != nil && err != io.EOF {
		f.Close()
		return nil, err
	}
	if bytes.IndexByte(head, 0) >= 0 {
		f.Close()
		return nil, nil
	}

	return struct {
		io.Reader
		io.Closer
	}{br, f}, nil
}
