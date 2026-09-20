package main

import (
	"errors"
	"fmt"
	"net/url"
)

const (
	apiBase  = "https://api.semanticscholar.org/graph/v1"
	fields   = "title,authors,year,publicationDate,venue,journal,citationCount,url"
	pageSize = 100
)

// getter performs an HTTP GET; the wasm build uses the PDK, tests inject a fake.
type getter func(url string, headers map[string]string) (status int, body []byte, err error)

// maxPages bounds pagination; 10 pages x 100 papers is far beyond any article_limit.
const maxPages = 10

// fetchAll pages through an author's papers, sorts them newest first, and
// keeps the first limit entries. All pages are read before truncating so
// "limit" means the newest N, not the API's first N.
func fetchAll(get getter, authorID, apiKey string, limit int) ([]Article, error) {
	if authorID == "" {
		return nil, errors.New("semantic scholar author id is empty")
	}
	headers := map[string]string{"Accept": "application/json"}
	if apiKey != "" {
		headers["x-api-key"] = apiKey
	}
	var out []Article
	offset := 0
	for page := 0; page < maxPages; page++ {
		u := fmt.Sprintf("%s/author/%s/papers?fields=%s&limit=%d&offset=%d", apiBase, url.PathEscape(authorID), fields, pageSize, offset)
		status, body, err := get(u, headers)
		if err != nil {
			return nil, err
		}
		if status != 200 {
			return nil, fmt.Errorf("semantic scholar returned HTTP %d: %s", status, truncate(string(body), 200))
		}
		p, err := parsePapersPage(body)
		if err != nil {
			return nil, fmt.Errorf("decode semantic scholar response: %w", err)
		}
		out = append(out, p.Articles...)
		if p.Next == nil || len(p.Articles) == 0 {
			break
		}
		offset = *p.Next
	}
	sortArticles(out)
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// truncate cuts s to at most n runes.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
