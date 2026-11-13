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
			folders["(без папки)"]++
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
			fa.Suggestion = "Пустая папка — можно удалить"
		} else if fa.IsShallow {
			fa.Suggestion = "Всего " + strings.TrimSpace(string([]rune{'0' + rune(count%10)})) + " закладки — можно слить с родительской"
		} else if depth > maxHealthyDepth {
			fa.Suggestion = "Слишком глубокая вложенность"
		}

		result = append(result, fa)
	}

	return result
}
