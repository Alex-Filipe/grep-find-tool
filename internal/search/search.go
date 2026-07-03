package search

import (
	"bufio"
	"context"
	"path/filepath"
	"sync"

	"github.com/user/grep-tool/internal/extract"
	"github.com/user/grep-tool/internal/matcher"
	"github.com/user/grep-tool/internal/walker"
)

// Result holds a single match or error found during search.
// If Err != nil, Path names the file that failed and LineNum/Line are zero.
type Result struct {
	Path    string
	LineNum int
	Line    string
	Err     error
}

// Search runs the full pipeline: walker -> extract -> matcher.
// Uses a worker pool for parallel processing over all given roots.
// Results are streamed incrementally through the returned channel (unordered).
// The returned error is only for setup failures (e.g. root doesn't exist);
// per-file errors are embedded in Result.Err.
func Search(ctx context.Context, roots []string, match matcher.MatchFunc, workers int) (<-chan Result, error) {
	if len(roots) == 0 {
		roots = []string{"."}
	}
	if workers < 1 {
		workers = 1
	}

	results := make(chan Result, workers*2)

	paths, err := mergePathChannels(ctx, roots)
	if err != nil {
		close(results)
		return results, err
	}

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for p := range paths {
				select {
				case <-ctx.Done():
					return
				default:
				}

				processFile(ctx, results, p, match)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results, nil
}

// mergePathChannels starts a walker for each root and merges all path channels.
// All walkers are started before any goroutine — if any root fails,
// no goroutines are leaked.
func mergePathChannels(ctx context.Context, roots []string) (<-chan string, error) {
	var walkers []<-chan string
	for _, root := range roots {
		ch, err := walker.Walk(ctx, root)
		if err != nil {
			return nil, err
		}
		walkers = append(walkers, ch)
	}

	out := make(chan string)
	go func() {
		defer close(out)
		var wg sync.WaitGroup
		for _, ch := range walkers {
			wg.Add(1)
			go func(ch <-chan string) {
				defer wg.Done()
				for p := range ch {
					select {
					case out <- p:
					case <-ctx.Done():
						return
					}
				}
			}(ch)
		}
		wg.Wait()
	}()

	return out, nil
}

// processFile extracts text from a file, scans it line by line,
// and sends each matching line as a Result to the channel.
func processFile(ctx context.Context, results chan<- Result, path string, match matcher.MatchFunc) {
	send := func(r Result) {
		select {
		case results <- r:
		case <-ctx.Done():
		}
	}

	ext := filepath.Ext(path)
	e := extract.For(ext)
	if e == nil {
		// No extractor for this extension — skip.
		return
	}

	r, err := e.Extract(path)
	if err != nil {
		send(Result{Path: path, Err: err})
		return
	}
	if r == nil {
		// Binary or unsupported — skip silently.
		return
	}
	defer r.Close()

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if match(line) {
			send(Result{
				Path:    path,
				LineNum: lineNum,
				Line:    line,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		send(Result{Path: path, Err: err})
	}
}
