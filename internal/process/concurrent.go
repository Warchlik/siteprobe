package process

import (
	"context"
	"sync"

	"siteprobe/internal/config"
	"siteprobe/internal/crawler"
)

func runConcurrent(
	ctx context.Context,
	c *crawler.Crawler,
	cfg *config.Config,
) <-chan crawler.Result {
	var workers int

	if cfg.MaxDepth == 0 {
		workers = min(cfg.Workers, len(cfg.URL))
	}

	out := make(chan crawler.Result)
	jobs := make(chan job)
	results := make(chan crawler.Result)

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for j := range jobs {
				res := c.Process(ctx, j.URL)
				res.Depth = j.Depth

				select {
				case results <- res:
				case <-ctx.Done():
					return
				}
			}
		})
	}

	go func() {
		defer close(out)
		defer wg.Wait()
		defer close(jobs)

		f := newFrontier(cfg.URL)
		inFlight := 0

		for !f.empty() || inFlight > 0 {
			var send chan<- job
			var next job
			if !f.empty() {
				send, next = jobs, f.peek()
			}

			select {
			case send <- next:
				f.pop()
				inFlight++

			case res := <-results:
				inFlight--

				select {
				case out <- res:
				case <-ctx.Done():
					return
				}

				f.expand(res, cfg.MaxDepth)

			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}
