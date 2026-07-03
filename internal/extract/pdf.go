package extract

import "io"

func init() {
	Register(".pdf", &pdfExtractor{})
}

type pdfExtractor struct{}

func (p *pdfExtractor) Extract(path string) (io.ReadCloser, error) {
	return nil, nil
}
