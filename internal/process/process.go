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
	"siteprobe/internal/queue"
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
		results = runConcurrent(
			ctx,
			c,
			cfg.URL,
			cfg.Workers,
		)
	} else if cfg.MaxDepth > 1 {
		results = queue.RunQueued(
			ctx,
			cfg.URL,
			c,
			cfg.MaxDepth,
		)
	} else {
		results = runSequential(
			ctx,
			c,
			cfg.URL,
		)
	}

	failed := report(results, cfg.Show)
	done := len(cfg.URL) - failed

	fmt.Printf("done: %d urls, %d failed, %d workers, %s\n",
		done, failed, max(cfg.Workers, 1), time.Since(start).Round(time.Millisecond))
	return nil
}

func report(results <-chan crawler.Result, show bool) (failed int) {
	for res := range results {
		if res.Err != nil {
			fmt.Printf("FAIL %s (%s): %v\n", res.URL, res.Took.Round(time.Millisecond), res.Err)
			failed++
			continue
		}
		fmt.Printf("OK %s (%s): %d links\n", res.URL, res.Took.Round(time.Millisecond), len(res.Links))

		if show {
			for _, link := range res.Links {
				fmt.Println("    ", link)
			}
		}
	}

	return failed
}
