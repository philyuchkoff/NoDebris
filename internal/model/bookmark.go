package model

import (
	"net/url"
	"strings"
	"time"
)

type Bookmark struct {
	ID                int
	Title             string
	URL               string
	Domain            string
	FolderPath        string
	DateAdded         time.Time
	StatusCode        int
	RedirectURL       string
	IsDuplicate       bool
	IsBroken          bool
	SuggestedCategory string
}

func (b *Bookmark) NormalizedURL() string {
	u, err := url.Parse(b.URL)
	if err != nil {
		return b.URL
	}
	u.Fragment = ""
	u.RawQuery = ""
	path := strings.TrimRight(u.Path, "/")
	u.Path = path
	result := u.String()
	result = strings.TrimSuffix(result, "/")
	return result
}

func (b *Bookmark) PopulateDomain() {
	u, err := url.Parse(b.URL)
	if err != nil {
		return
	}
	b.Domain = strings.TrimPrefix(u.Host, "www.")
}

type FolderAnalysis struct {
	Path        string
	Count       int
	Depth       int
	IsEmpty     bool
	IsShallow   bool
	Suggestion  string
	Children    []FolderAnalysis
}
