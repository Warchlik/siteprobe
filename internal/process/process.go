// Package process will implement standard structure for 2 implementation,
// with contains synchonicous crawling and gorutines runtime
package process

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"siteprobe/internal/config"
	"siteprobe/internal/crawler"
	"siteprobe/internal/fetcher"
	"siteprobe/internal/parser"
)

const requestTimeout = 10 * time.Second

func Run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	c := crawler.NewCrawler(
		fetcher.NewFetcher(requestTimeout, cfg.UserAgent),
		parser.NewHTMLParser(),
	)

	start := time.Now()

	var results <-chan crawler.Result

	if cfg.Workers > 1 {
		results = runConcurrent(ctx, c, cfg)
	} else {
		results = runSequential(ctx, c, cfg)
	}

	processed, failed := report(results, cfg.Show)

	total := fmt.Sprint(processed)
	if cfg.MaxDepth == 0 {
		total = fmt.Sprintf("%d of %d", processed, len(cfg.URL))
	}

	fmt.Printf("done: %s urls, %d failed, %d workers, %s\n",
		total, failed, max(cfg.Workers, 1), time.Since(start).Round(time.Millisecond))
	return nil
}

func report(results <-chan crawler.Result, show bool) (processed, failed int) {
	for res := range results {
		processed++

		if res.Err != nil {
			fmt.Printf("FAIL [d%d] %s (%s): %v\n", res.Depth, res.URL, res.Took.Round(time.Millisecond), res.Err)
			failed++
			continue
		}
		fmt.Printf("OK [d%d] %s (%s): %d links\n", res.Depth, res.URL, res.Took.Round(time.Millisecond), len(res.Links))

		if show {
			for _, link := range res.Links {
				fmt.Println("    ", link)
			}
		}
	}

	return processed, failed
}
