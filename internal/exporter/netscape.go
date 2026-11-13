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
	type treeNode struct {
		node     *folderNode
		children map[string]*treeNode
	}

	root := &treeNode{
		node:     &folderNode{Name: ""},
		children: make(map[string]*treeNode),
	}

	for path, bms := range folderMap {
		parts := strings.Split(path, "/")
		current := root

		for i, part := range parts {
			if _, ok := current.children[part]; !ok {
				current.children[part] = &treeNode{
					node:     &folderNode{Name: part},
					children: make(map[string]*treeNode),
				}
			}
			current = current.children[part]

			if i == len(parts)-1 {
				current.node.Bookmarks = append(current.node.Bookmarks, bms...)
			}
		}
	}

	var buildList func(tn *treeNode) []*folderNode
	buildList = func(tn *treeNode) []*folderNode {
		var result []*folderNode
		for _, child := range tn.children {
			child.node.Children = buildList(child)
			result = append(result, child.node)
		}
		return result
	}

	return buildList(root)
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
