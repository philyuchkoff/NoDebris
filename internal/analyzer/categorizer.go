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
	{Domain: "github.com", Category: "Dev/GitHub"},
	{Domain: "gitlab.com", Category: "Dev/GitLab"},
	{Domain: "stackoverflow.com", Category: "Dev/Stack Overflow"},
	{Domain: "habr.com", Category: "Articles"},
	{Domain: "medium.com", Category: "Articles"},
	{Domain: "youtube.com", Category: "Video"},
	{Domain: "reddit.com", Category: "Social"},
	{Domain: "twitter.com", Category: "Social"},
	{Domain: "x.com", Category: "Social"},
	{Domain: "linkedin.com", Category: "Social"},
	{Domain: "docs.google.com", Category: "Documents"},
	{Domain: "notion.so", Category: "Documents"},
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
