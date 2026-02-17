package utils

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

type GitHubCommit struct {
	SHA string `json:"sha"`
}

func FetchUpdates() error {
	log.Info().Msg("Fetching latest repo state from GitHub")

	client := &http.Client{}
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", owner, repo, branch),
		nil,
	)
	req.Header.Set("User-Agent", "iceslab-config-setup-script")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var commit GitHubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return fmt.Errorf("failed to decode GitHub API response: %w", err)
	}
	remoteSHA := commit.SHA

	// Read name of parent dir for local SHA
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	parentDir := filepath.Dir(execPath)
	localSHA := []byte(strings.TrimPrefix(filepath.Base(parentDir), "iceslab-"))
	switch localSHA {
	case nil:
		log.Info().Msg("No local commit SHA found; treating as first run")
	default:
		if string(localSHA) == remoteSHA {
			log.Info().Msg("Repo is already up to date (remote: " + remoteSHA + ", local: " + string(localSHA) + ")")
			return nil
		}
	}

	log.Info().Str("remote_sha", remoteSHA).Msg("New commit found; updating local repo state")

	newDir := filepath.Join("iceslab-" + remoteSHA)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("failed to create new directory for updated repo state: %w", err)
	}

	tarReq, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", owner, repo, branch),
		nil,
	)
	tarReq.Header.Set("User-Agent", "iceslab-config-setup-script")

	tarResp, err := client.Do(tarReq)
	if err != nil {
		return fmt.Errorf("failed to download repo tarball: %w", err)
	}
	defer tarResp.Body.Close()

	if tarResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(tarResp.Body)
		return fmt.Errorf("GitHub tarball request failed with status %d: %s", tarResp.StatusCode, string(body))
	}

	if err := extractTarGz(tarResp.Body, newDir); err != nil {
		return fmt.Errorf("failed to extract repo tarball: %w", err)
	}

	if err := os.WriteFile(".last_commit", []byte(remoteSHA), 0644); err != nil {
		return fmt.Errorf("failed to write last commit SHA: %w", err)
	}

	iceslabPath := filepath.Join(newDir, "iceslab")
	if err := os.Chmod(iceslabPath, 0755); err != nil {
		return fmt.Errorf("failed to chmod iceslab: %w", err)
	}

	log.Info().Msg("Repo state updated successfully")

	return nil
}

func extractTarGz(gzipStream io.Reader, dest string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// GitHub tarballs have a root folder (owner-repo-sha/).
		// We strip it to extract contents directly into our SHA folder.
		parts := strings.Split(header.Name, "/")
		if len(parts) <= 1 {
			continue
		}
		relPath := filepath.Join(parts[1:]...)
		target := filepath.Join(dest, relPath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}
