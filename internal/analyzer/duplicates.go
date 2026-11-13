package analyzer

import "github.com/philyuchkoff/NoDebris/internal/model"

type DupGroup struct {
	URL       string
	Bookmarks []model.Bookmark
}

func FindDuplicates(bookmarks []model.Bookmark) []DupGroup {
	groups := make(map[string][]model.Bookmark)
	for _, b := range bookmarks {
		key := b.NormalizedURL()
		groups[key] = append(groups[key], b)
	}

	var result []DupGroup
	for _, group := range groups {
		if len(group) > 1 {
			result = append(result, DupGroup{
				URL:       group[0].URL,
				Bookmarks: group,
			})
		}
	}
	return result
}
