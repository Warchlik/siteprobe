package process

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/goleak"

	"siteprobe/internal/config"
	"siteprobe/internal/crawler"
	"siteprobe/internal/parser"
)

type runFunc func(context.Context, *crawler.Crawler, *config.Config) <-chan crawler.Result

type fakeFetcher struct{}

func (fakeFetcher) Fetch(ctx context.Context, rawURL string) ([]byte, error) {
	select {
	case <-time.After(5 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var n int
	if _, err := fmt.Sscanf(rawURL, "http://site.test/p%d", &n); err != nil {
		return nil, err
	}

	return fmt.Appendf(nil, `<a href="/p%d">l</a><a href="/p%d">r</a>`, 2*n, 2*n+1), nil
}

func testCrawler() *crawler.Crawler {
	return crawler.NewCrawler(fakeFetcher{}, parser.NewHTMLParser())
}

// testConfig crawls 4 levels from /p1: 1+2+4+8+16 = 31 pages.
func testConfig(workers int) *config.Config {
	return &config.Config{
		URL:      []string{"http://site.test/p1"},
		Workers:  workers,
		MaxDepth: 4,
	}
}

func TestNoLeakWhenDrained(t *testing.T) {
	modes := map[string]runFunc{
		"sequential": runSequential,
		"concurrent": runConcurrent,
	}

	for name, run := range modes {
		t.Run(name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			pages := 0
			for range run(context.Background(), testCrawler(), testConfig(4)) {
				pages++
			}

			if pages != 31 {
				t.Fatalf("pages = %d, want 31", pages)
			}
		})
	}
}

func TestNoLeakWhenCancelledEarly(t *testing.T) {
	modes := map[string]runFunc{
		"sequential": runSequential,
		"concurrent": runConcurrent,
	}

	for name, run := range modes {
		t.Run(name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			ctx, cancel := context.WithCancel(context.Background())
			out := run(ctx, testCrawler(), testConfig(4))

			<-out
			cancel()
		})
	}
}
