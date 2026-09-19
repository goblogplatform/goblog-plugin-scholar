package main

import (
	"encoding/json"
	"html"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Article is one publication as rendered on the Research page.
type Article struct {
	Title     string `json:"title"`
	Authors   string `json:"authors"`
	URL       string `json:"url"`
	Year      int    `json:"year"`
	Date      string `json:"date"` // publicationDate, YYYY-MM-DD when known
	Journal   string `json:"journal"`
	Citations int    `json:"citations"`
}

type papersPage struct {
	Next     *int
	Articles []Article
}

// parsePapersPage decodes one page of /author/{id}/papers.
func parsePapersPage(b []byte) (papersPage, error) {
	var raw struct {
		Next *int `json:"next"`
		Data []struct {
			URL             string `json:"url"`
			Title           string `json:"title"`
			Year            int    `json:"year"`
			PublicationDate string `json:"publicationDate"`
			Venue           string `json:"venue"`
			CitationCount   int    `json:"citationCount"`
			Authors         []struct {
				Name string `json:"name"`
			} `json:"authors"`
			Journal *struct {
				Name string `json:"name"`
			} `json:"journal"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return papersPage{}, err
	}
	page := papersPage{Next: raw.Next}
	for _, p := range raw.Data {
		names := make([]string, 0, len(p.Authors))
		for _, a := range p.Authors {
			if a.Name != "" {
				names = append(names, a.Name)
			}
		}
		journal := p.Venue
		if p.Journal != nil && p.Journal.Name != "" {
			journal = p.Journal.Name
		}
		page.Articles = append(page.Articles, Article{
			Title: p.Title, Authors: strings.Join(names, ", "), URL: p.URL, Year: p.Year,
			Date: p.PublicationDate, Journal: journal, Citations: p.CitationCount,
		})
	}
	return page, nil
}

// sortArticles orders newest first: year, then publication date, then citations.
func sortArticles(as []Article) {
	sort.SliceStable(as, func(i, j int) bool {
		if as[i].Year != as[j].Year {
			return as[i].Year > as[j].Year
		}
		if as[i].Date != as[j].Date {
			return as[i].Date > as[j].Date
		}
		return as[i].Citations > as[j].Citations
	})
}

// safeHref allows only absolute http(s) URLs into href attributes.
func safeHref(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	return html.EscapeString(u.String())
}

// renderArticlesHTML is the same markup goblog's compiled-in scholar plugin produced.
func renderArticlesHTML(articles []Article) string {
	if len(articles) == 0 {
		return `<p>No publications found.</p>`
	}
	var b strings.Builder
	for _, a := range articles {
		b.WriteString(`<div style="margin-bottom: 12px; padding-bottom: 12px; border-bottom: 1px solid #eee;">`)
		if href := safeHref(a.URL); href != "" {
			b.WriteString(`<div><a href="` + href + `">` + html.EscapeString(a.Title) + `</a></div>`)
		} else {
			b.WriteString(`<div>` + html.EscapeString(a.Title) + `</div>`)
		}
		if a.Authors != "" {
			b.WriteString(`<div style="color: #666; font-size: 13px;">` + html.EscapeString(a.Authors) + `</div>`)
		}
		var meta []string
		if a.Year > 0 {
			meta = append(meta, strconv.Itoa(a.Year))
		}
		if a.Journal != "" {
			meta = append(meta, html.EscapeString(a.Journal))
		}
		if a.Citations > 0 {
			meta = append(meta, strconv.Itoa(a.Citations)+" citations")
		}
		if len(meta) > 0 {
			b.WriteString(`<div style="color: #888; font-size: 13px;">` + strings.Join(meta, " &middot; ") + `</div>`)
		}
		b.WriteString(`</div>`)
	}
	return b.String()
}

// cachedArticles is what the plugin keeps under the store key "articles".
type cachedArticles struct {
	FetchedAt time.Time `json:"fetched_at"`
	Articles  []Article `json:"articles"`
}

func (c cachedArticles) fresh(now time.Time, cacheHours int) bool {
	if c.FetchedAt.IsZero() {
		return false
	}
	return now.Sub(c.FetchedAt) < time.Duration(cacheHours)*time.Hour
}

// parseIntSetting reads a positive integer setting with a default.
func parseIntSetting(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return def
	}
	return n
}

const (
	unavailableHTML = `<div class="alert alert-warning" role="alert">Publications are temporarily unavailable. Please check back later.</div>`
	noIDHTML        = `<div class="alert alert-warning" role="alert">Semantic Scholar Author ID not configured. Set it in the Scholar Publications plugin settings.</div>`
)
