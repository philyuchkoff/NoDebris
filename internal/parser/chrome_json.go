package parser

import (
	"encoding/json"
	"io"
	"time"

	"github.com/philyuchkoff/NoDebris/internal/model"
)

type chromeBookmarkJSON struct {
	Name     string               `json:"name"`
	URL      string               `json:"url"`
	DateAdded string              `json:"date_added"`
	Children []chromeBookmarkJSON `json:"children"`
}

type chromeRoot struct {
	Roots struct {
		BookmarkBar chromeBookmarkJSON `json:"bookmark_bar"`
		Other       chromeBookmarkJSON `json:"other"`
		Synced      chromeBookmarkJSON `json:"synced"`
	} `json:"roots"`
}

var rootNames = map[string]bool{
	"Bookmark Bar":    true,
	"Other Bookmarks": true,
	"Mobile Bookmarks": true,
}

func ParseChromeJSON(r io.Reader) ([]model.Bookmark, error) {
	var root chromeRoot
	if err := json.NewDecoder(r).Decode(&root); err != nil {
		return nil, err
	}

	var bookmarks []model.Bookmark
	id := 0

	var collect func(node chromeBookmarkJSON, folderPath string)
	collect = func(node chromeBookmarkJSON, folderPath string) {
		isFolder := node.URL == "" && len(node.Children) > 0
		skipRoot := rootNames[node.Name]

		currentPath := folderPath
		if isFolder && !skipRoot {
			if currentPath == "" {
				currentPath = node.Name
			} else {
				currentPath = currentPath + "/" + node.Name
			}
		}

		if node.URL != "" {
			var t time.Time
			if node.DateAdded != "" {
				var ms int64
				for _, c := range node.DateAdded {
					if c >= '0' && c <= '9' {
						ms = ms*10 + int64(c-'0')
					}
				}
				if ms > 0 {
					t = time.UnixMicro(ms)
				}
			}

			b := model.Bookmark{
				ID:         id,
				Title:      node.Name,
				URL:        node.URL,
				FolderPath: folderPath,
				DateAdded:  t,
			}
			b.PopulateDomain()
			bookmarks = append(bookmarks, b)
			id++
		}

		for _, child := range node.Children {
			collect(child, currentPath)
		}
	}

	collect(root.Roots.BookmarkBar, "")
	collect(root.Roots.Other, "")
	collect(root.Roots.Synced, "")

	return bookmarks, nil
}
