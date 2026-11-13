package analyzer

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/philyuchkoff/NoDebris/internal/model"
)

type CategoryRule struct {
	Domain   string `json:"domain"`
	Category string `json:"category"`
}

var defaultRules = []CategoryRule{
	{Domain: "github.com", Category: "Разработка/GitHub"},
	{Domain: "gitlab.com", Category: "Разработка/GitLab"},
	{Domain: "stackoverflow.com", Category: "Разработка/Stack Overflow"},
	{Domain: "habr.com", Category: "Статьи"},
	{Domain: "medium.com", Category: "Статьи"},
	{Domain: "youtube.com", Category: "Видео"},
	{Domain: "reddit.com", Category: "Соцсети"},
	{Domain: "twitter.com", Category: "Соцсети"},
	{Domain: "x.com", Category: "Соцсети"},
	{Domain: "linkedin.com", Category: "Соцсети"},
	{Domain: "docs.google.com", Category: "Документы"},
	{Domain: "notion.so", Category: "Документы"},
}

func LoadRules(path string) ([]CategoryRule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultRules, nil
	}
	var rules []CategoryRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return defaultRules, nil
	}
	return rules, nil
}

func Categorize(bookmarks []model.Bookmark, rules []CategoryRule) []model.Bookmark {
	result := make([]model.Bookmark, len(bookmarks))
	copy(result, bookmarks)

	for i := range result {
		for _, rule := range rules {
			if strings.Contains(result[i].Domain, rule.Domain) {
				result[i].SuggestedCategory = rule.Category
				break
			}
		}
	}

	return result
}
