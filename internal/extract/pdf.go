package extract

import (
	"io"
)

func init() {
	Register(".pdf", &pdfExtractor{})
}

type pdfExtractor struct{}

func (p *pdfExtractor) Extract(path string) (io.ReadCloser, error) {
	// TODO: implement PDF text extraction
	return nil, nil
}
