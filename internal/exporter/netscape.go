package exporter

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/philyuchkoff/NoDebris/internal/model"
)

func ExportNetscape(w io.Writer, bookmarks []model.Bookmark) error {
	_, err := fmt.Fprint(w, `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
`)
	if err != nil {
		return err
	}

	folderMap := make(map[string][]model.Bookmark)
	var rootBookmarks []model.Bookmark

	for _, b := range bookmarks {
		if b.FolderPath == "" {
			rootBookmarks = append(rootBookmarks, b)
		} else {
			folderMap[b.FolderPath] = append(folderMap[b.FolderPath], b)
		}
	}

	tree := buildFolderTree(folderMap)
	writeFolder(w, tree, 1)
	writeBookmarks(w, rootBookmarks, 2)

	_, err = fmt.Fprint(w, "</DL><p>\n")
	return err
}

type folderNode struct {
	Name      string
	Bookmarks []model.Bookmark
	Children  []*folderNode
}

func buildFolderTree(folderMap map[string][]model.Bookmark) []*folderNode {
	root := make(map[string]*folderNode)

	for path, bms := range folderMap {
		parts := strings.Split(path, "/")
		accum := ""
		var parent map[string]*folderNode
		parent = root

		for i, part := range parts {
			if accum == "" {
				accum = part
			} else {
				accum = accum + "/" + part
			}

			if _, ok := parent[accum]; !ok {
				parent[accum] = &folderNode{Name: part}
			}
			node := parent[accum]

			if i == len(parts)-1 {
				node.Bookmarks = append(node.Bookmarks, bms...)
			}

			if len(parts) > 1 && i < len(parts)-1 {
				if node.Children == nil {
					node.Children = make([]*folderNode, 0)
				}
			}

			parent = root
			_ = node
		}
	}

	seen := make(map[string]bool)
	var result []*folderNode
	for path, node := range root {
		parts := strings.Split(path, "/")
		if len(parts) == 1 && !seen[path] {
			seen[path] = true
			result = append(result, node)
		}
	}
	return result
}

func writeFolder(w io.Writer, nodes []*folderNode, indent int) {
	for _, node := range nodes {
		pad := strings.Repeat("    ", indent)
		fmt.Fprintf(w, "%s<DT><H3>%s</H3>\n%s<DL><p>\n", pad, html.EscapeString(node.Name), pad)
		writeBookmarks(w, node.Bookmarks, indent+1)
		writeFolder(w, node.Children, indent+1)
		fmt.Fprintf(w, "%s</DL><p>\n", pad)
	}
}

func writeBookmarks(w io.Writer, bookmarks []model.Bookmark, indent int) {
	pad := strings.Repeat("    ", indent)
	for _, b := range bookmarks {
		fmt.Fprintf(w, "%s<DT><A HREF=\"%s\">%s</A>\n",
			pad, html.EscapeString(b.URL), html.EscapeString(b.Title))
	}
}
