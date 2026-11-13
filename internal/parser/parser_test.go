package parser

import (
	"strings"
	"testing"
)

func TestParseNetscape(t *testing.T) {
	input := `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
    <DT><H3 ADD_DATE="1700000000">Dev</H3>
    <DL><p>
        <DT><A HREF="https://github.com" ADD_DATE="1700000001">GitHub</A>
        <DT><A HREF="https://github.com" ADD_DATE="1700000002">GitHub (duplicate)</A>
        <DT><A HREF="https://go.dev" ADD_DATE="1700000003">Go</A>
        <DT><H3 ADD_DATE="1700000004">Golang</H3>
        <DL><p>
            <DT><A HREF="https://pkg.go.dev" ADD_DATE="1700000005">Go Packages</A>
        </DL><p>
    </DL><p>
    <DT><A HREF="https://example.com">No Folder</A>
</DL><p>`

	bookmarks, err := ParseNetscape(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bookmarks) != 5 {
		t.Fatalf("expected 5 bookmarks, got %d", len(bookmarks))
	}

	if bookmarks[0].FolderPath != "Dev" {
		t.Errorf("expected folder 'Dev', got '%s'", bookmarks[0].FolderPath)
	}

	if bookmarks[3].FolderPath != "Dev/Golang" {
		t.Errorf("expected folder 'Dev/Golang', got '%s'", bookmarks[3].FolderPath)
	}

	if bookmarks[4].FolderPath != "" {
		t.Errorf("expected empty folder path, got '%s'", bookmarks[4].FolderPath)
	}

	if bookmarks[0].Domain != "github.com" {
		t.Errorf("expected domain 'github.com', got '%s'", bookmarks[0].Domain)
	}
}

func TestParseChromeJSON(t *testing.T) {
	input := `{
  "roots": {
    "bookmark_bar": {
      "name": "Bookmark Bar",
      "children": [
        {
          "name": "GitHub",
          "url": "https://github.com",
          "date_added": "13400000000000000"
        },
        {
          "name": "Dev",
          "children": [
            {
              "name": "Go",
              "url": "https://go.dev",
              "date_added": "13400000000000001"
            }
          ]
        }
      ]
    },
    "other": {
      "name": "Other Bookmarks",
      "children": []
    },
    "synced": {
      "name": "Mobile Bookmarks",
      "children": []
    }
  }
}`

	bookmarks, err := ParseChromeJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bookmarks) != 2 {
		t.Fatalf("expected 2 bookmarks, got %d", len(bookmarks))
	}

	if bookmarks[0].URL != "https://github.com" {
		t.Errorf("expected github URL, got '%s'", bookmarks[0].URL)
	}

	if bookmarks[1].FolderPath != "Dev" {
		t.Errorf("expected folder 'Dev', got '%s'", bookmarks[1].FolderPath)
	}
}
