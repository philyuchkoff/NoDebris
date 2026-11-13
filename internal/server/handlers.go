package server

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/philyuchkoff/NoDebris/internal/analyzer"
	"github.com/philyuchkoff/NoDebris/internal/exporter"
	"github.com/philyuchkoff/NoDebris/internal/model"
	"github.com/philyuchkoff/NoDebris/internal/parser"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "base.html", map[string]interface{}{
		"Title":   "NoDebris",
		"Content": "index_page",
	})
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20)
	file, header, err := r.FormFile("bookmarks")
	var data []model.Bookmark

	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		switch ext {
		case ".html", ".htm":
			data, err = parser.ParseNetscape(file)
		case ".json":
			data, err = parser.ParseChromeJSON(file)
		default:
			http.Error(w, "Unsupported file format", http.StatusBadRequest)
			return
		}
	} else {
		path := r.FormValue("path")
		if path == "" {
			http.Error(w, "No file uploaded and no path provided", http.StatusBadRequest)
			return
		}
		data, err = loadFromPath(path)
	}

	if err != nil {
		http.Error(w, "Parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	dups := analyzer.FindDuplicates(data)
	dupIDs := make(map[int]bool)
	for _, g := range dups {
		for _, b := range g.Bookmarks {
			dupIDs[b.ID] = true
		}
	}
	for i := range data {
		if dupIDs[data[i].ID] {
			data[i].IsDuplicate = true
		}
	}

	mu.Lock()
	bookmarks = data
	mu.Unlock()

	http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
}

func loadFromPath(path string) ([]model.Bookmark, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if strings.HasSuffix(path, ".json") {
		return parser.ParseChromeJSON(f)
	}
	return parser.ParseNetscape(f)
}

func handleBookmarks(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	if len(bookmarks) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	filter := r.URL.Query().Get("filter")
	search := r.URL.Query().Get("search")
	folder := r.URL.Query().Get("folder")

	filtered := filterBookmarks(bookmarks, filter, search, folder)

	tmpl.ExecuteTemplate(w, "base.html", map[string]interface{}{
		"Title":     "Bookmarks",
		"Content":   "bookmarks_page",
		"Bookmarks": filtered,
		"Filter":    filter,
		"Search":    search,
		"Folder":    folder,
		"Folders":   uniqueFolders(bookmarks),
		"Total":     len(bookmarks),
		"Shown":     len(filtered),
	})
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["ids"]

	mu.Lock()
	toRemove := make(map[int]bool)
	for _, idStr := range ids {
		id, _ := strconv.Atoi(idStr)
		toRemove[id] = true
	}

	var kept []model.Bookmark
	for _, b := range bookmarks {
		if !toRemove[b.ID] {
			kept = append(kept, b)
		}
	}
	bookmarks = kept
	mu.Unlock()

	http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
}

func handleMove(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ids := r.Form["ids"]
	targetFolder := r.FormValue("folder")

	mu.Lock()
	moveMap := make(map[int]bool)
	for _, idStr := range ids {
		id, _ := strconv.Atoi(idStr)
		moveMap[id] = true
	}

	for i := range bookmarks {
		if moveMap[bookmarks[i].ID] {
			bookmarks[i].FolderPath = targetFolder
		}
	}
	mu.Unlock()

	http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
}

func handleDuplicates(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	dups := analyzer.FindDuplicates(bookmarks)
	mu.RUnlock()

	tmpl.ExecuteTemplate(w, "base.html", map[string]interface{}{
		"Title":     "Duplicates",
		"Content":   "duplicates_page",
		"DupGroups": dups,
		"TotalDups": len(dups),
	})
}

func handleDupResolve(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	deleteIDs := r.Form["ids"]

	deleteSet := make(map[int]bool)
	for _, idStr := range deleteIDs {
		id, _ := strconv.Atoi(idStr)
		deleteSet[id] = true
	}

	if len(deleteSet) == 0 {
		http.Redirect(w, r, "/duplicates", http.StatusSeeOther)
		return
	}

	mu.Lock()
	dups := analyzer.FindDuplicates(bookmarks)

	removeSet := make(map[int]bool)
	for _, group := range dups {
		marked := 0
		total := len(group.Bookmarks)
		for _, b := range group.Bookmarks {
			if deleteSet[b.ID] {
				removeSet[b.ID] = true
				marked++
			}
		}
		if marked >= total {
			delete(removeSet, group.Bookmarks[0].ID)
		}
	}

	var kept []model.Bookmark
	for _, b := range bookmarks {
		if !removeSet[b.ID] {
			kept = append(kept, b)
		}
	}
	bookmarks = kept
	mu.Unlock()

	http.Redirect(w, r, "/duplicates", http.StatusSeeOther)
}

func handleStructure(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	analysis := analyzer.AnalyzeStructure(bookmarks)
	total := len(bookmarks)
	mu.RUnlock()

	tmpl.ExecuteTemplate(w, "base.html", map[string]interface{}{
		"Title":     "Folder Structure",
		"Content":   "structure_page",
		"Analysis":  analysis,
		"TotalBm":   total,
		"TotalFold": len(analysis),
	})
}

func handleCategorize(w http.ResponseWriter, r *http.Request) {
	rules, _ := analyzer.LoadRules("categories.json")
	mu.Lock()
	bookmarks = analyzer.Categorize(bookmarks, rules)
	mu.Unlock()
	http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
}

func handleCheckLinksStart(w http.ResponseWriter, r *http.Request) {
	checkMu.Lock()
	if checkRunning {
		checkMu.Unlock()
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<p style=\"color: var(--pico-color-slate-500);\">Check is already running...</p>"))
		return
	}
	checkRunning = true
	checkDone = 0
	checkResults = nil

	mu.RLock()
	toCheck := make([]model.Bookmark, len(bookmarks))
	copy(toCheck, bookmarks)
	mu.RUnlock()
	checkTotal = len(toCheck)

	stopCh := make(chan struct{})
	checkStop = stopCh
	checkMu.Unlock()

	go func() {
		defer func() {
			checkMu.Lock()
			checkRunning = false
			checkMu.Unlock()
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			select {
			case <-stopCh:
				cancel()
			case <-ctx.Done():
			}
		}()

		resultMap := make(map[int]analyzer.LinkResult)
		for res := range analyzer.CheckLinks(ctx, toCheck, 10) {
			checkMu.Lock()
			checkResults = append(checkResults, res)
			resultMap[res.ID] = res
			checkDone++
			checkMu.Unlock()
		}

		mu.Lock()
		for i := range bookmarks {
			if res, ok := resultMap[bookmarks[i].ID]; ok {
				bookmarks[i].StatusCode = res.StatusCode
				if res.Error != "" {
					bookmarks[i].IsBroken = true
				}
				if res.StatusCode >= 400 {
					bookmarks[i].IsBroken = true
				}
				if res.StatusCode >= 300 && res.StatusCode < 400 {
					bookmarks[i].RedirectURL = "redirect"
				}
			}
		}
		mu.Unlock()
	}()

	w.Header().Set("HX-Trigger", "poll-progress")
	tmpl.ExecuteTemplate(w, "check_progress.html", map[string]interface{}{
		"Done":    0,
		"Total":   checkTotal,
		"Running": true,
	})
}

func handleCheckLinksProgress(w http.ResponseWriter, r *http.Request) {
	checkMu.Lock()
	done := checkDone
	total := checkTotal
	results := make([]analyzer.LinkResult, len(checkResults))
	copy(results, checkResults)
	running := checkRunning
	checkMu.Unlock()

	if running && done < total {
		w.Header().Set("HX-Trigger", "poll-progress")
	}

	tmpl.ExecuteTemplate(w, "check_progress.html", map[string]interface{}{
		"Done":    done,
		"Total":   total,
		"Results": results,
		"Running": running,
	})
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	var buf bytes.Buffer
	exporter.ExportNetscape(&buf, bookmarks)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=bookmarks_export.html")
	buf.WriteTo(w)
}

func uniqueFolders(list []model.Bookmark) []string {
	seen := make(map[string]bool)
	var result []string
	for _, b := range list {
		if b.FolderPath != "" && !seen[b.FolderPath] {
			seen[b.FolderPath] = true
			result = append(result, b.FolderPath)
		}
	}
	return result
}

func filterBookmarks(list []model.Bookmark, filter, search, folder string) []model.Bookmark {
	var result []model.Bookmark
	for _, b := range list {
		if folder != "" && b.FolderPath != folder {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(b.Title), strings.ToLower(search)) &&
			!strings.Contains(strings.ToLower(b.URL), strings.ToLower(search)) {
			continue
		}
		switch filter {
		case "broken":
			if !b.IsBroken {
				continue
			}
		case "redirect":
			if b.RedirectURL == "" {
				continue
			}
		case "duplicate":
			if !b.IsDuplicate {
				continue
			}
		}
		result = append(result, b)
	}
	return result
}
