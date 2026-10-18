// Package crawler scans pages and collects their links.
package crawler

import (
	"context"
	"fmt"
	"net/url"
	"time"

	fetcher "siteprobe/internal/fetcher"
	parser "siteprobe/internal/parser"
)

type Result struct {
	URL   string
	Links []string
	Err   error
	Took  time.Duration
	Depth int
}

type Crawler struct {
	fetcher fetcher.LinkFetcher
	parser  parser.LinkExtractor
}

func NewCrawler(f fetcher.LinkFetcher, p parser.LinkExtractor) *Crawler {
	return &Crawler{
		fetcher: f,
		parser:  p,
	}
}

func (c *Crawler) Process(ctx context.Context, rawURL string) (res Result) {
	res.URL = rawURL
	start := time.Now()
	defer func() { res.Took = time.Since(start) }()

	base, err := url.Parse(rawURL)
	if err != nil || base.Host == "" {
		res.Err = fmt.Errorf("invalid url %q: use a full address, e.g. https://example.com", rawURL)
		return res
	}

	body, err := c.fetcher.Fetch(ctx, rawURL)
	if err != nil {
		res.Err = err
		return res
	}

	res.Links, res.Err = c.parser.Extract(base, body)
	return res
}
