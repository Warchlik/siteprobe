// Package parser extracts same-domain links from HTML documents.
package parser

import (
	"bytes"
	"errors"
	"io"
	"net/url"
	"path"
	"slices"
	"strings"

	html "golang.org/x/net/html"
)

type LinkExtractor interface {
	Extract(base *url.URL, body []byte) ([]string, error)
}

type HTMLParser struct {
	skippedExtensions []string
}

// Compile-time check that HTMLParser implements LinkExtractor.
var _ LinkExtractor = (*HTMLParser)(nil)

func NewHTMLParser() *HTMLParser {
	return &HTMLParser{
		skippedExtensions: []string{
			".jpg",
			".jpeg",
			".png",
			".gif",
			".svg",
			".pdf",
			".zip",
			".css",
			".js",
			".ico",
			".doc",
			".docx",
			".xlsx",
			".xls",
			".wasm",
		},
	}
}

func (hp *HTMLParser) Extract(base *url.URL, body []byte) ([]string, error) {
	z := html.NewTokenizer(bytes.NewReader(body))
	var links []string

	for {
		switch z.Next() {
		case html.ErrorToken:
			if err := z.Err(); !errors.Is(err, io.EOF) {
				return links, err
			}
			return links, nil

		case html.StartTagToken, html.SelfClosingTagToken:
			tok := z.Token()
			if tok.Data != "a" {
				continue
			}
			for _, attr := range tok.Attr {
				if attr.Key != "href" {
					continue
				}
				if link, ok := hp.normalize(base, attr.Val); ok {
					links = append(links, link)
				}
			}
		}
	}
}

func (hp *HTMLParser) normalize(base *url.URL, href string) (string, bool) {
	href = strings.TrimSpace(href)
	if href == "" {
		return "", false
	}

	ref, err := url.Parse(href)
	if err != nil {
		return "", false
	}

	url := base.ResolveReference(ref)

	if url.Scheme != "http" && url.Scheme != "https" {
		return "", false
	}
	if url.Host != base.Host {
		return "", false
	}
	if slices.Contains(hp.skippedExtensions, strings.ToLower(path.Ext(url.Path))) {
		return "", false
	}

	url.Fragment = ""
	return url.String(), true
}
