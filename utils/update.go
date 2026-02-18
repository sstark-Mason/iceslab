package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// https://api.github.com/repos/sstark-mason/iceslab/releases/latest
// https://api.github.com/repos/sstark-mason/iceslab/releases/tags/bookmarks-latest

const (
	owner  = "sstark-mason"
	repo   = "iceslab"
	branch = "main"
	// bookmarksURL = "https://api.github.com/repos/sstark-mason/iceslab/zipball/bookmarks-latest" // This is actually full source (repo)
	bookmarksURL = "https://github.com/sstark-mason/iceslab/releases/download/bookmarks-latest/bookmarks.zip"
) // curl -L -o bookmarks.zip https://github.com/sstark-mason/iceslab/releases/download/bookmarks-latest/bookmarks.zip

type Client struct {
	http    *http.Client
	baseURL string
	token   string
}

func NewClient(token string) *Client {
	return &Client{
		http:    &http.Client{},
		baseURL: "https://api.github.com",
		token:   token,
	}
}

func (c *Client) Update0() error {
	shouldWeUpdate, err := c.CompareRemoteManifestETag()
	if err != nil {
		return err
	}
	if !shouldWeUpdate {
		log.Info().Msg("No updates available; skipping update process")
		return nil
	}

	zipData, err := c.DownloadRepoZip()
	if err != nil {
		return err
	}
	targets := []string{"iceslab", "assets"}

	err = unzipRepoZip0(zipData, targets)
	if err != nil {
		return err
	}

	log.Info().Msg("Update process completed successfully")
	return nil
}

func (c *Client) Update() error {
	shouldWeUpdate, err := c.CompareRemoteManifestETag()
	if err != nil {
		return err
	}
	if !shouldWeUpdate {
		log.Info().Msg("No updates available; skipping update process")
		return nil
	}

	zipData, err := c.DownloadRepoZip()
	if err != nil {
		return err
	}
	targets := []string{"iceslab", "assets"}

	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "iceslab-update-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	// Do not defer remove; let the shell command handle cleanup

	err = unzipRepoZip(zipData, targets, tempDir)
	if err != nil {
		os.RemoveAll(tempDir) // Clean up on error
		return err
	}

	// Schedule file moves after process exit to avoid "text file busy"
	cmd := exec.Command("sh", "-c", fmt.Sprintf("sleep 1 && cp -r %s/* . && rm -rf %s", tempDir, tempDir))
	err = cmd.Start()
	if err != nil {
		os.RemoveAll(tempDir) // Clean up on error
		return fmt.Errorf("failed to schedule file moves: %w", err)
	}

	log.Info().Msg("Update process initiated; files will be replaced after exit")
	return nil
}

func (c *Client) DownloadRepoZip() ([]byte, error) {
	// resp, err := c.http.Get(fmt.Sprintf("https://github.com/%s/%s/zipball/%s", owner, repo, branch))
	resp, err := c.http.Get(fmt.Sprintf("https://github.com/%s/%s/archive/%s.zip", owner, repo, branch))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download repo zip: status %d", resp.StatusCode)
	}

	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo zip data: %w", err)
	}

	return zipData, nil
}

func unzipRepoZip0(zipData []byte, targets []string) error {
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, f := range zr.File {
		fpath := f.Name

		// Split path and get relative path after top-level dir (e.g., "iceslab-main/")
		parts := strings.Split(fpath, "/")
		if len(parts) < 2 {
			continue // Skip invalid paths or top-level dir itself
		}
		relativePath := strings.Join(parts[1:], "/")

		// Check if the relative path matches any target (prefix match)
		matchesTarget := false
		for _, target := range targets {
			if strings.HasPrefix(relativePath, target+"/") || relativePath == target {
				matchesTarget = true
				break
			}
		}
		if !matchesTarget {
			continue // Skip files not matching targets
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(relativePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(relativePath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory for file: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}
		defer rc.Close()

		data, err := io.ReadAll(rc)
		if err != nil {
			return fmt.Errorf("failed to read file data from zip: %w", err)
		}

		err = os.WriteFile(relativePath, data, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to write file from zip: %w", err)
		}
	}

	return nil
}

func unzipRepoZip(zipData []byte, targets []string, destDir string) error {
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, f := range zr.File {
		fpath := f.Name

		// Split path and get relative path after top-level dir (e.g., "iceslab-main/")
		parts := strings.Split(fpath, "/")
		if len(parts) < 2 {
			continue // Skip invalid paths or top-level dir itself
		}
		relativePath := strings.Join(parts[1:], "/")

		// Check if the relative path matches any target (prefix match)
		matchesTarget := false
		for _, target := range targets {
			if strings.HasPrefix(relativePath, target+"/") || relativePath == target {
				matchesTarget = true
				break
			}
		}
		if !matchesTarget {
			continue // Skip files not matching targets
		}

		// Build full path in temp dir
		fullPath := filepath.Join(destDir, relativePath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fullPath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory for file: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}
		defer rc.Close()

		data, err := io.ReadAll(rc)
		if err != nil {
			return fmt.Errorf("failed to read file data from zip: %w", err)
		}

		err = os.WriteFile(fullPath, data, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to write file from zip: %w", err)
		}
	}

	return nil
}

func moveFiles(srcDir, destDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destDir, relPath)
		if info.IsDir() {
			return os.MkdirAll(destPath, os.ModePerm)
		}
		// Always copy the file (fallback for cross-device moves)
		if err := copyFile(path, destPath); err != nil {
			return err
		}
		return os.Remove(path)
	})
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Preserve file mode
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

func (c *Client) FetchManifestETag() (string, error) {

	req, _ := http.NewRequest(
		"HEAD",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/manifest.yaml", owner, repo),
		nil,
	)
	req.Header.Set("User-Agent", "iceslab-script")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch manifest ETag: status %d", resp.StatusCode)
	}

	return resp.Header.Get("ETag"), nil
}
