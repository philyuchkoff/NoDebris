package exporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/philyuchkoff/NoDebris/internal/model"
)

func TestExportNetscape_PreservesNestedFolders(t *testing.T) {
	bookmarks := []model.Bookmark{
		{ID: 0, Title: "Root", URL: "https://root.com", FolderPath: ""},
		{ID: 1, Title: "GH", URL: "https://github.com", FolderPath: "Dev/GitHub"},
		{ID: 2, Title: "GL", URL: "https://gitlab.com", FolderPath: "Dev/GitLab"},
		{ID: 3, Title: "GoDoc", URL: "https://pkg.go.dev", FolderPath: "Dev/GitHub/Go"},
	}

	var buf bytes.Buffer
	err := ExportNetscape(&buf, bookmarks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := buf.String()

	if !strings.Contains(result, "https://root.com") {
		t.Error("root bookmark missing")
	}
	if !strings.Contains(result, "https://github.com") {
		t.Error("Dev/GitHub bookmark missing")
	}
	if !strings.Contains(result, "https://gitlab.com") {
		t.Error("Dev/GitLab bookmark missing")
	}
	if !strings.Contains(result, "https://pkg.go.dev") {
		t.Error("Dev/GitHub/Go bookmark missing")
	}

	if !strings.Contains(result, "<H3>Dev</H3>") {
		t.Error("Dev folder missing")
	}
	if !strings.Contains(result, "<H3>GitHub</H3>") {
		t.Error("GitHub subfolder missing")
	}
	if !strings.Contains(result, "<H3>Go</H3>") {
		t.Error("Go subfolder missing")
	}
}
