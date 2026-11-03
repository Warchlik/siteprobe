package process

import (
	"errors"
	"net/url"
	"strings"

	"siteprobe/internal/crawler"
)

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
		link, err := url.Parse(u)
		if err != nil {
			return nil
		}
		f.seen[link.String()] = struct{}{}
		f.queue = append(f.queue, job{URL: u})
	}

	return f
}

func (f *frontier) expand(res crawler.Result, maxDepth int) {
	if res.Err != nil || res.Depth >= maxDepth {
		return
	}

	for _, link := range res.Links {
		linkNorm, err := normalizeURL(link)
		if err != nil {
			continue
		}

		if _, isSeen := f.seen[linkNorm]; isSeen {
			continue
		}

		f.seen[linkNorm] = struct{}{}
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

func normalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(strings.ToLower(rawURL))
	if err != nil {
		return "", err
	}

	if parsedURL.Host == "" {
		return "", errors.New("URL do not have a host")
	}

	return parsedURL.String(), nil
}
