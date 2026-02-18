package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog/log"
	"go.yaml.in/yaml/v4"
)

type Bookmark struct {
	Name string `yaml:"name"`
	URL  any    `yaml:"url"`
}

func (b *Bookmark) GetURL(stationNum string) (Bookmark, error) {
	// TODO: Rewrite this function.
	switch url := b.URL.(type) {
	case string:
		return Bookmark{Name: b.Name, URL: url}, nil
	case map[any]any:
		// Try lookup with stationNum as string
		if urlVal, ok := url[stationNum]; ok {
			if urlStr, ok := urlVal.(string); ok {
				return Bookmark{Name: b.Name, URL: urlStr}, nil
			}
		}
		// If not found, try parsing stationNum to int (for YAML keys like 01 parsed as 1)
		if stationInt, err := strconv.Atoi(stationNum); err == nil {
			if urlVal, ok := url[stationInt]; ok {
				if urlStr, ok := urlVal.(string); ok {
					return Bookmark{Name: b.Name, URL: urlStr}, nil
				}
			}
		}
	case []any:
		// Try to index by stationNum
		if stationInt, err := strconv.Atoi(stationNum); err == nil && stationInt > 0 {
			index := stationInt - 1
			if index >= 0 && index < len(url) {
				if urlStr, ok := url[index].(string); ok {
					return Bookmark{Name: b.Name, URL: urlStr}, nil
				}
			}
		}

		// Fall back to the first URL in the array if stationNum isn't found or if it's an array
		if len(url) > 0 {
			if urlStr, ok := url[0].(string); ok {
				return Bookmark{Name: b.Name, URL: urlStr}, nil
			}
		}
	}
	return Bookmark{}, fmt.Errorf("invalid bookmark URL format for %s", b.Name)
}

func CollectBookmarks(dir string, stationNum string) ([]Bookmark, error) {
	var bookmarks []Bookmark
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subBookmarks, err := CollectBookmarks(filepath.Join(dir, entry.Name()), stationNum)
			if err != nil {
				log.Warn().Err(err).Str("directory", entry.Name()).Msg("Failed to collect bookmarks from subdirectory")
				continue
			}
			bookmarks = append(bookmarks, subBookmarks...)
		} else {
			// Only process files, not directories
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				log.Warn().Err(err).Str("file", entry.Name()).Msg("Failed to read bookmark file")
				continue
			}
			var bm Bookmark
			err = yaml.Unmarshal(data, &bm)
			if err != nil {
				log.Warn().Err(err).Str("file", entry.Name()).Msg("Failed to parse bookmark file")
				continue
			}
			finalBM, err := bm.GetURL(stationNum)
			if err != nil {
				log.Warn().Err(err).Str("file", entry.Name()).Msg("Failed to get bookmark URL for station")
				continue
			}
			bookmarks = append(bookmarks, finalBM)
		}
	}

	for _, bm := range bookmarks {
		log.Debug().Str("name", bm.Name).Str("url", fmt.Sprintf("%v", bm.URL)).Msg("Collected bookmark")
	}

	return bookmarks, nil
}

func InsertBookmarks(browser string, path string, bookmarks []Bookmark) error {
	// Read the existing policies.json file
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Unmarshal into a map
	var policies map[string]any
	if err := json.Unmarshal(data, &policies); err != nil {
		return err
	}

	// Insert or update the bookmarks
	switch browser {
	case "firefox":
		policies["policies"].(map[string]any)["ManagedBookmarks"] = bookmarks
	case "chromium":
		policies["ManagedBookmarks"] = bookmarks
	}

	// Marshal back to JSON
	updatedData, err := json.MarshalIndent(policies, "", "  ")
	if err != nil {
		return err
	}

	// Write back to file
	return os.WriteFile(path, updatedData, 0644)
}
