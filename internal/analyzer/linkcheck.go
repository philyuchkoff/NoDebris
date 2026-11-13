package analyzer

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/philyuchkoff/NoDebris/internal/model"
)

const defaultConcurrency = 10

type LinkResult struct {
	ID         int
	StatusCode int
	Error      string
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func CheckLinks(ctx context.Context, bookmarks []model.Bookmark, concurrency int) <-chan LinkResult {
	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}

	results := make(chan LinkResult, len(bookmarks))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, b := range bookmarks {
		wg.Add(1)
		go func(bm model.Bookmark) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			status, errStr := checkOne(ctx, bm.URL)
			select {
			case results <- LinkResult{ID: bm.ID, StatusCode: status, Error: errStr}:
			case <-ctx.Done():
			}
		}(b)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func checkOne(ctx context.Context, url string) (int, string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, err.Error()
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return 0, "cancelled"
		}

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if reqErr != nil {
			return 0, reqErr.Error()
		}
		resp, err = httpClient.Do(req)
		if err != nil {
			return 0, err.Error()
		}
		defer resp.Body.Close()
		return resp.StatusCode, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return resp.StatusCode, ""
	}

	return resp.StatusCode, ""
}
