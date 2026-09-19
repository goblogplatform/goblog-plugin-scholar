package main

import (
	"strings"
	"testing"
	"time"
)

const pageJSON = `{"offset":0,"next":2,"data":[
 {"paperId":"p1","url":"https://www.semanticscholar.org/paper/p1","title":"Old & Cited","year":2019,"publicationDate":"2019-03-01","venue":"Conf A","journal":{"name":"J. A"},"citationCount":40,"authors":[{"name":"A. One"},{"name":"B. Two"}]},
 {"paperId":"p2","url":"javascript:alert(1)","title":"<Newest>","year":2024,"publicationDate":"2024-06-01","venue":"","journal":null,"citationCount":1,"authors":[{"name":"C. Three"}]}
]}`

func TestParsePapersPage(t *testing.T) {
	page, err := parsePapersPage([]byte(pageJSON))
	if err != nil {
		t.Fatal(err)
	}
	if page.Next == nil || *page.Next != 2 || len(page.Articles) != 2 {
		t.Fatalf("page = %+v", page)
	}
	a := page.Articles[0]
	if a.Title != "Old & Cited" || a.Authors != "A. One, B. Two" || a.Year != 2019 || a.Journal != "J. A" || a.Citations != 40 || a.URL != "https://www.semanticscholar.org/paper/p1" || a.Date != "2019-03-01" {
		t.Errorf("article = %+v", a)
	}
	if page.Articles[1].Journal != "" {
		t.Errorf("no journal/venue should be empty, got %q", page.Articles[1].Journal)
	}
	if _, err := parsePapersPage([]byte("{")); err == nil {
		t.Error("bad JSON should error")
	}
}

func TestSortArticles(t *testing.T) {
	as := []Article{
		{Title: "b", Year: 2020, Date: "2020-01-01", Citations: 5},
		{Title: "c", Year: 2021, Date: "2021-01-01", Citations: 1},
		{Title: "a", Year: 2020, Date: "2020-05-01", Citations: 2},
		{Title: "d", Year: 2020, Date: "2020-05-01", Citations: 9},
	}
	sortArticles(as)
	got := as[0].Title + as[1].Title + as[2].Title + as[3].Title
	if got != "cdab" {
		t.Errorf("order = %s, want cdab (year desc, date desc, citations desc)", got)
	}
}

func TestRenderArticlesHTML(t *testing.T) {
	if got := renderArticlesHTML(nil); got != "<p>No publications found.</p>" {
		t.Errorf("empty = %q", got)
	}
	page, _ := parsePapersPage([]byte(pageJSON))
	html := renderArticlesHTML(page.Articles)
	for _, want := range []string{`href="https://www.semanticscholar.org/paper/p1"`, "Old &amp; Cited", "A. One, B. Two", "2019 &middot; J. A &middot; 40 citations", "&lt;Newest&gt;", "2024 &middot; 1 citations"} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in\n%s", want, html)
		}
	}
	if strings.Contains(html, "javascript:") {
		t.Error("unsafe URL must not be linked")
	}
	if safeHref("ftp://x") != "" || safeHref("https://ok.test/a?b=1") != "https://ok.test/a?b=1" {
		t.Error("safeHref rules")
	}
}

func TestCacheFreshness(t *testing.T) {
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	c := cachedArticles{FetchedAt: now.Add(-2 * time.Hour)}
	if !c.fresh(now, 24) || c.fresh(now, 1) {
		t.Error("freshness by cache_hours")
	}
	if (cachedArticles{}).fresh(now, 24) {
		t.Error("zero FetchedAt is never fresh")
	}
	if parseIntSetting("", 50) != 50 || parseIntSetting("7", 50) != 7 || parseIntSetting("x", 50) != 50 || parseIntSetting("0", 50) != 50 {
		t.Error("parseIntSetting")
	}
}
