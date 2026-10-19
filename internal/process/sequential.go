package process

import (
	"context"

	"siteprobe/internal/config"
	"siteprobe/internal/crawler"
)

func runSequential(
	ctx context.Context,
	c *crawler.Crawler,
	cfg *config.Config,
) <-chan crawler.Result {
	out := make(chan crawler.Result)

	go func() {
		defer close(out)

		f := newFrontier(cfg.URL)

		for !f.empty() {
			j := f.pop()

			res := c.Process(ctx, j.URL)
			res.Depth = j.Depth

			select {
			case out <- res:
			case <-ctx.Done():
				return
			}

			f.expand(res, cfg.MaxDepth)
		}
	}()

	return out
}
