package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type rawWatchlistItemResponse struct {
	TMDBID    *int64 `json:"tmdbId"`
	MediaType string `json:"mediaType"`
	Title     string `json:"title"`
	Name      string `json:"name"`
	Overview  string `json:"overview"`
}

type rawWatchlistPageResponse struct {
	Page         *int                       `json:"page"`
	TotalPages   *int                       `json:"totalPages"`
	TotalResults *int                       `json:"totalResults"`
	Results      []rawWatchlistItemResponse `json:"results"`
}

func authenticatedUserID(ctx context.Context, client *APIClient) (int64, error) {
	res, err := client.Request(ctx, "GET", "/api/v1/auth/me", "", nil)
	if err != nil {
		return 0, fmt.Errorf("request current authenticated user: %w", err)
	}
	if !StatusIsOK(res.StatusCode) {
		return 0, fmt.Errorf("request current authenticated user: status %d: %s", res.StatusCode, string(res.Body))
	}
	var user struct {
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(res.Body, &user); err != nil {
		return 0, fmt.Errorf("parse current authenticated user: %w", err)
	}
	var id int64
	if err := json.Unmarshal(user.ID, &id); err != nil {
		var s string
		if string(user.ID) == "" || json.Unmarshal(user.ID, &s) != nil {
			return 0, fmt.Errorf("parse current authenticated user: id must be a positive integer")
		}
		parsed, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse current authenticated user id %q: %w", s, err)
		}
		id = parsed
	}
	if id <= 0 {
		return 0, fmt.Errorf("parse current authenticated user: id must be a positive integer")
	}
	return id, nil
}

// fetchWatchlist enumerates the complete upstream user watchlist before returning
// it, so callers can distinguish a truly absent resource from an incomplete read.
func fetchWatchlist(ctx context.Context, client *APIClient, userID int64) ([]rawWatchlistItemResponse, int64, error) {
	const maxWatchlistPages = 500
	items := make([]rawWatchlistItemResponse, 0)
	var total int64 = -1
	var totalPages int
	for page := 1; page <= maxWatchlistPages; page++ {
		endpoint := fmt.Sprintf("/api/v1/user/%d/watchlist?page=%d", userID, page)
		res, err := client.Request(ctx, "GET", endpoint, "", nil)
		if err != nil {
			return nil, 0, fmt.Errorf("read watchlist page %d: %w", page, err)
		}
		if !StatusIsOK(res.StatusCode) {
			return nil, 0, fmt.Errorf("read watchlist page %d: status %d: %s", page, res.StatusCode, string(res.Body))
		}
		var raw rawWatchlistPageResponse
		if err := json.Unmarshal(res.Body, &raw); err != nil {
			return nil, 0, fmt.Errorf("parse watchlist page %d: %w", page, err)
		}
		if raw.Results == nil || raw.Page == nil || *raw.Page != page || raw.TotalPages == nil || raw.TotalResults == nil || *raw.TotalPages < 0 || *raw.TotalResults < 0 {
			return nil, 0, fmt.Errorf("parse watchlist page %d: response is missing valid page, totalPages, totalResults, or results", page)
		}
		if page == 1 {
			total = int64(*raw.TotalResults)
			totalPages = *raw.TotalPages
			// Seerr returns totalPages=0 for some empty Plex results.
			if total == 0 {
				if len(raw.Results) != 0 {
					return nil, 0, fmt.Errorf("parse watchlist page 1: empty totalResults with non-empty results")
				}
				if totalPages != 0 && totalPages != 1 {
					return nil, 0, fmt.Errorf("parse watchlist page 1: inconsistent empty pagination metadata")
				}
				return items, total, nil
			}
			if totalPages < 1 || totalPages > maxWatchlistPages {
				return nil, 0, fmt.Errorf("parse watchlist page 1: invalid totalPages %d", totalPages)
			}
		} else if int64(*raw.TotalResults) != total || *raw.TotalPages != totalPages {
			return nil, 0, fmt.Errorf("parse watchlist page %d: pagination metadata changed during enumeration", page)
		}
		for _, item := range raw.Results {
			if item.TMDBID == nil || *item.TMDBID <= 0 || (item.MediaType != "movie" && item.MediaType != "tv") {
				return nil, 0, fmt.Errorf("parse watchlist page %d: item has invalid tmdbId or mediaType", page)
			}
		}
		items = append(items, raw.Results...)
		if page == totalPages {
			if int64(len(items)) != total {
				return nil, 0, fmt.Errorf("read watchlist: received %d items but upstream reports %d", len(items), total)
			}
			return items, total, nil
		}
	}
	return nil, 0, fmt.Errorf("read watchlist: exceeded maximum page count")
}
