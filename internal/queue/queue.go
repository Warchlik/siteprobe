// Package queue
package queue

import (
	"context"

	"siteprobe/internal/crawler"
)

type job struct {
	URL   string
	Depth int
}

func RunQueued(
	ctx context.Context,
	startURLs []string,
	c *crawler.Crawler,
	maxDepth int,
) <-chan crawler.Result {
	out := make(chan crawler.Result)

	go func() {
		defer close(out)

		visited := make(map[string]struct{}, len(startURLs))

		queue := make([]job, 0, len(startURLs))

		for _, u := range startURLs {
			visited[u] = struct{}{}
			queue = append(queue, job{URL: u})
		}

		for len(queue) > 0 {
			j := queue[0]
			queue = queue[1:]

			res := c.Process(ctx, j.URL)
			res.Depth = j.Depth

			select {
			case out <- res:
			case <-ctx.Done():
				return
			}

			if res.Err != nil && j.Depth >= maxDepth {
				continue
			}

			for _, link := range res.Links {
				if _, seen := visited[link]; seen {
					continue
				}

				visited[link] = struct{}{}
				queue = append(queue, job{URL: link, Depth: j.Depth + 1})
			}
		}
	}()

	return out
}
