package main

import (
	"fmt"
	"os"

	"github.com/philyuchkoff/NoDebris/internal/analyzer"
	"github.com/philyuchkoff/NoDebris/internal/parser"
	"github.com/philyuchkoff/NoDebris/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("NoDebris — Chrome bookmark cleaner")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  serve              start web interface")
		fmt.Println("  analyze <file>     analyze bookmarks in terminal")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "serve":
		fmt.Println("NoDebris running at http://localhost:8080")
		if err := server.Run(":8080"); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}

	case "analyze":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Specify file path: analyze <file>")
			os.Exit(1)
		}
		runAnalyze(os.Args[2])

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runAnalyze(path string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	bookmarks, err := parser.ParseNetscape(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total bookmarks: %d\n", len(bookmarks))

	dups := analyzer.FindDuplicates(bookmarks)
	fmt.Printf("Duplicate groups: %d\n", len(dups))
	for _, g := range dups {
		fmt.Printf("  %s (%d copies)\n", g.URL, len(g.Bookmarks))
	}

	analysis := analyzer.AnalyzeStructure(bookmarks)
	problems := 0
	for _, a := range analysis {
		if a.Suggestion != "" {
			problems++
		}
	}
	fmt.Printf("Folders with issues: %d\n", problems)
	for _, a := range analysis {
		if a.Suggestion != "" {
			fmt.Printf("  %s: %s\n", a.Path, a.Suggestion)
		}
	}

	rules, err := analyzer.LoadRules("categories.json")
	if err == nil {
		categorized := analyzer.Categorize(bookmarks, rules)
		withCat := 0
		for _, b := range categorized {
			if b.SuggestedCategory != "" {
				withCat++
			}
		}
		fmt.Printf("Bookmarks with suggested category: %d\n", withCat)
	}
}
