package parser

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/philyuchkoff/NoDebris/internal/model"
)

var (
	reLink    = regexp.MustCompile(`(?i)<A\s+[^>]*HREF="([^"]*)"[^>]*>(.*?)</A>`)
	reHref    = regexp.MustCompile(`(?i)HREF="([^"]*)"`)
	reAddDate = regexp.MustCompile(`(?i)ADD_DATE="(\d+)"`)
	reFolder  = regexp.MustCompile(`(?i)<H3[^>]*>(.*?)</H3>`)
)

func ParseNetscape(r io.Reader) ([]model.Bookmark, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var bookmarks []model.Bookmark
	id := 0
	folderStack := []string{}

	for _, rawLine := range lines {
		lowerLine := strings.ToLower(rawLine)

		if strings.Contains(lowerLine, "<h3") {
			match := reFolder.FindStringSubmatch(rawLine)
			if len(match) >= 2 {
				name := extractAnchorText(match[1])
				folderStack = append(folderStack, name)
			}
			continue
		}

		if strings.Contains(lowerLine, "<a ") && strings.Contains(lowerLine, "href=") {
			hrefMatch := reHref.FindStringSubmatch(rawLine)
			addDateMatch := reAddDate.FindStringSubmatch(rawLine)

			linkMatch := reLink.FindStringSubmatch(rawLine)
			var title string
			if len(linkMatch) >= 3 {
				title = extractAnchorText(linkMatch[2])
			}

			if len(hrefMatch) >= 2 {
				var t time.Time
				if len(addDateMatch) >= 2 {
					if unix, err := parseUnix(addDateMatch[1]); err == nil && unix > 0 {
						t = time.Unix(unix, 0)
					}
				}

				b := model.Bookmark{
					ID:         id,
					Title:      title,
					URL:        hrefMatch[1],
					FolderPath: strings.Join(folderStack, "/"),
					DateAdded:  t,
				}
				b.PopulateDomain()
				bookmarks = append(bookmarks, b)
				id++
			}
			continue
		}

		if strings.Contains(lowerLine, "</dl>") {
			if len(folderStack) > 0 {
				folderStack = folderStack[:len(folderStack)-1]
			}
		}
	}

	return bookmarks, nil
}

func parseUnix(s string) (int64, error) {
	var result int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		result = result*10 + int64(c-'0')
	}
	return result, nil
}

func extractAnchorText(s string) string {
	s = strings.TrimSpace(s)
	for {
		start := strings.Index(s, "<")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], ">")
		if end == -1 {
			break
		}
		end += start
		s = s[:start] + s[end+1:]
	}
	return strings.TrimSpace(s)
}
