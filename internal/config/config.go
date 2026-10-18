// Package config
package config

import (
	"errors"
	"flag"
	"slices"
	"strings"
	"unicode"
)

type Config struct {
	URL       []string
	Workers   int
	MaxDepth  int
	Rate      int
	UserAgent string
	Show      bool
}

func LoadConfig() (*Config, error) {
	var cfg Config
	var urlValues string

	flag.StringVar(&urlValues, "url", "", "Crawling url")
	flag.IntVar(&cfg.Workers, "workers", 1, "Flag definites how many workers will connect and search in website")
	flag.IntVar(&cfg.MaxDepth, "max-depth", 1, "Flag definites how deep worker can enter int to website definie by links")
	flag.IntVar(&cfg.Rate, "rate", 1, "Flag define rate")
	flag.StringVar(&cfg.UserAgent, "user-agent", "Crawler-Agent/default", "Flag defines custom user agnet")
	flag.BoolVar(&cfg.Show, "show", false, "Show collected links form url/urls")
	flag.Parse()

	urlArr, err := parseURLList(urlValues)
	if err != nil {
		return nil, err
	}
	cfg.URL = urlArr

	if cfg.MaxDepth <= 0 {
		return nil, errors.New("MaxDepth can not be less then 0")
	}

	if cfg.Workers < 1 {
		return nil, errors.New("worker can not be less then 0")
	}

	return &cfg, nil
}

func parseURLList(raw string) ([]string, error) {
	urls := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	slices.Sort(urls)
	urls = slices.Compact(urls)

	if len(urls) == 0 {
		return nil, errors.New("url is required")
	}

	return urls, nil
}
