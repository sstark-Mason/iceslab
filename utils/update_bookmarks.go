package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

const (
	pathBookmarks        = "assets/bookmarks/"
	pathFirefoxPolicies  = "assets/etc/firefox/policies/policies.json"
	pathChromiumPolicies = "assets/etc/chromium/policies/managed/policies.json"
)

func InstallBookmarks(stationID string) error {
	log.Info().Msg("Installing bookmarks")

	bookmarks, err := CollectBookmarks("assets/bookmarks/", stationID)
	if err != nil {
		return fmt.Errorf("failed to collect bookmarks: %w", err)
	}

	err = InsertBookmarks("firefox", pathFirefoxPolicies, bookmarks)
	if err != nil {
		return fmt.Errorf("failed to insert firefox bookmarks: %w", err)
	}

	err = InsertBookmarks("chromium", pathChromiumPolicies, bookmarks)
	if err != nil {
		return fmt.Errorf("failed to insert chromium bookmarks: %w", err)
	}

	log.Info().Msg("Bookmarks installed successfully")
	return nil

}

func (c *Client) UpdateBookmarks() error {
	localETagBytes, err := os.ReadFile(".etag_bookmarks")
	localETag := ""
	if err == nil {
		localETag = string(localETagBytes)
		log.Debug().Str("local_bookmarks_etag", localETag).Msg("Read local bookmarks ETag")
	} else {
		log.Info().Msg("No local bookmarks ETag found; treating as first run")
	}

	zipData, latestETag, err := c.FetchLatestBookmarks(localETag)
	if err != nil {
		return fmt.Errorf("failed to fetch latest bookmarks: %w", err)
	}

	if zipData == nil {
		log.Info().Msg("Bookmarks are up to date; no update needed")
		return nil
	}

	pathOldBookmarks := "assets/bookmarks_old/"
	backedUpOldBookmarks := false

	if _, err := os.Stat(pathBookmarks); err == nil {
		log.Debug().Msg("Existing bookmarks directory found; backing up before update")

		err = MoveFile(pathBookmarks, pathOldBookmarks)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to move existing bookmarks to backup directory")
		}
		backedUpOldBookmarks = true
	}

	err = unzipInto("assets/bookmarks/", zipData)
	if err != nil {
		if backedUpOldBookmarks {
			log.Warn().Err(err).Msg("Failed to unzip new bookmarks; attempting to restore old bookmarks from backup")
			restoreErr := MoveFile(pathOldBookmarks, pathBookmarks)
			if restoreErr != nil {
				log.Warn().Err(restoreErr).Msg("Failed to restore old bookmarks after unzip failure")
			} else {
				log.Info().Msg("Restored old bookmarks after unzip failure")
			}
		}
		return fmt.Errorf("failed to unzip bookmarks: %w", err)
	}

	// Clean up any old bookmarks files that might be left over from previous versions
	if backedUpOldBookmarks {
		err = os.RemoveAll(pathOldBookmarks)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to clean up old bookmarks backup directory")
		}
	}

	err = os.WriteFile(".etag_bookmarks", []byte(latestETag), 0644)
	if err != nil {
		return fmt.Errorf("failed to save latest bookmarks ETag: %w", err)
	}

	log.Info().Str("latest_bookmarks_etag", latestETag).Msg("Bookmarks updated and ETag saved locally")
	return nil
}

func (c *Client) FetchLatestBookmarks(localETag string) ([]byte, string, error) {
	// Returns bookmarks zip data, latest ETag, error.
	latestETag := localETag
	request, err := http.NewRequest("GET", latestBookmarksReleaseURL, nil)
	if err != nil {
		return nil, latestETag, err
	}

	if localETag != "" {
		request.Header.Set("If-None-Match", localETag)
	}

	response, err := c.http.Do(request)
	if err != nil {
		return nil, latestETag, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusNotModified:
		return nil, latestETag, nil
	case http.StatusOK:
		zipData, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, latestETag, fmt.Errorf("failed to read bookmarks zip data: %w", err)
		}
		latestETag = response.Header.Get("ETag")
		return zipData, latestETag, nil
	default:
		return nil, latestETag, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}
}
