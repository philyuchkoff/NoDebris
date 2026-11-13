package analyzer

import (
	"strings"

	"github.com/philyuchkoff/NoDebris/internal/model"
)

const maxHealthyDepth = 3

func AnalyzeStructure(bookmarks []model.Bookmark) []model.FolderAnalysis {
	folders := make(map[string]int)
	depthMap := make(map[string]int)

	for _, b := range bookmarks {
		if b.FolderPath == "" {
			folders["(no folder)"]++
			continue
		}
		parts := strings.Split(b.FolderPath, "/")
		accum := ""
		for i, part := range parts {
			if accum == "" {
				accum = part
			} else {
				accum = accum + "/" + part
			}
			folders[accum]++
			depthMap[accum] = i + 1
		}
	}

	var result []model.FolderAnalysis
	for path, count := range folders {
		depth := depthMap[path]
		fa := model.FolderAnalysis{
			Path:      path,
			Count:     count,
			Depth:     depth,
			IsEmpty:   count == 0,
			IsShallow: count <= 2 && depth > 0,
		}

		if fa.IsEmpty {
			fa.Suggestion = "Empty folder — can be removed"
		} else if fa.IsShallow {
			fa.Suggestion = "Only " + itoa(count) + " bookmark(s) — consider merging with parent"
		} else if depth > maxHealthyDepth {
			fa.Suggestion = "Too deeply nested"
		}

		result = append(result, fa)
	}

	return result
}

func itoa(n int) string {
	if n < 0 {
		return "0"
	}
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
