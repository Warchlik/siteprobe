package process

import "siteprobe/internal/crawler"

type job struct {
	URL   string
	Depth int
}

type frontier struct {
	seen  map[string]struct{}
	queue []job
}

func newFrontier(startURLs []string) *frontier {
	f := &frontier{
		seen:  make(map[string]struct{}, len(startURLs)),
		queue: make([]job, 0, len(startURLs)),
	}

	for _, u := range startURLs {
		f.seen[u] = struct{}{}
		f.queue = append(f.queue, job{URL: u})
	}

	return f
}

func (f *frontier) expand(res crawler.Result, maxDepth int) {
	if res.Err != nil || res.Depth >= maxDepth {
		return
	}

	for _, link := range res.Links {
		if _, isSeen := f.seen[link]; isSeen {
			continue
		}

		f.seen[link] = struct{}{}
		f.queue = append(f.queue, job{URL: link, Depth: res.Depth + 1})
	}
}

func (f *frontier) empty() bool {
	return len(f.queue) == 0
}

func (f *frontier) peek() job {
	return f.queue[0]
}

func (f *frontier) pop() job {
	j := f.queue[0]
	f.queue = f.queue[1:]
	return j
}
