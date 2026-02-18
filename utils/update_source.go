package utils

import (
	"fmt"
	"io"
	"net/http"
)

func (c *Client) UpdateSource() error {
	zipData, err := c.DownloadSourceZip()
	if err != nil {
		return fmt.Errorf("failed to download source zip: %w", err)
	}

	err = unzipInto("", zipData)
	if err != nil {
		return fmt.Errorf("failed to unzip source: %w", err)
	}

	return nil
}

func (c *Client) DownloadSourceZip() ([]byte, error) {
	response, err := c.http.Get(fmt.Sprintf("https://github.com/%s/%s/zipball/%s", owner, repo, branch))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download repo zip: status %d", response.StatusCode)
	}

	zipData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo zip data: %w", err)
	}

	return zipData, nil
}
