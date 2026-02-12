package main

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.yaml.in/yaml/v4"
)

//go:embed all:assets
var embedded embed.FS

const (
	owner  = "sstark-mason"
	repo   = "iceslab"
	branch = "main"
)

type GitHubCommit struct {
	SHA string `json:"sha"`
}

func dumpAssets(src, dest string) error {
	entries, err := embedded.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fullSrc := filepath.Join(src, entry.Name())
		fullDest := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			err = os.MkdirAll(fullDest, 0755)
			if err != nil {
				return err
			}
			err = dumpAssets(fullSrc, fullDest)
			if err != nil {
				return err
			}
		} else {
			alreadyExists := false
			if _, err := os.Stat(fullDest); err == nil {
				alreadyExists = true
				log.Debug().Str("file", entry.Name()).Str("path", fullDest).Msg("Asset already exists, skipping")
				continue
			}
			if !alreadyExists {
				data, err := embedded.ReadFile(fullSrc)
				if err != nil {
					log.Warn().Err(err).Str("file", entry.Name()).Str("path", fullDest).Msg("Failed to read asset")
					continue
				}
				err = writeFile(fullDest, data, 0644)
				if err != nil {
					log.Warn().Err(err).Str("file", entry.Name()).Str("path", fullDest).Msg("Failed to write asset")
					continue
				} else {
					log.Info().Str("file", entry.Name()).Str("path", fullDest).Msg("Asset written")
					continue
				}
			}
		}
	}

	return nil
}

func writeFile(path string, data []byte, perm os.FileMode) error {
	log.Debug().Str("path", path).Msg("Writing file")
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}

func promptForStationNumber() (string, error) {
	var stationNum string
	log.Info().Msg("Please enter the station number (e.g., 01, 02, etc.):")
	_, err := fmt.Scanln(&stationNum)
	if err != nil {
		return "", err
	}
	if len(stationNum) == 1 && stationNum >= "1" && stationNum <= "9" {
		stationNum = "0" + stationNum
	}
	return stationNum, nil
}

type Bookmark struct {
	Name string `yaml:"name"`
	URL  any    `yaml:"url"`
}

func (b *Bookmark) GetURL(stationNum string) (Bookmark, error) {
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

func insertBookmarks(browser string, path string, bookmarks []Bookmark) error {
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

func fetchUpdates() error {
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

func getAssetPath() string {
	// Check for local "assets" directory first
	if _, err := os.Stat("assets"); err == nil {
		log.Debug().Msg("Using local assets path")
		return "assets"
	}
	// Fallback to embedded assets
	log.Debug().Msg("Using embedded assets path")
	return "embedded"
}

func runScript(path string) error {
	if _, err := os.Stat(path); err == nil {
		log.Info().Str("script", path).Msg("Running script")
		err = os.Chmod(path, 0755)
		if err != nil {
			return fmt.Errorf("failed to make script executable: %w", err)
		}
		cmd := fmt.Sprintf("./%s", path)
		if err := exec.Command(cmd).Run(); err != nil {
			return fmt.Errorf("failed to run script: %w", err)
		}
		log.Info().Str("script", path).Msg("Script completed successfully")
	}
	return nil
}

func main() {

	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	var err error

	update := flag.Bool("update", false, "Fetch latest repo state from GitHub before setup")
	flag.Parse()

	if *update {
		err = fetchUpdates()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to update repo state")
		}
		return
	}

	stationNum, err := promptForStationNumber()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read station number")
	}

	log.Info().Str("station_number", stationNum).Msg("Starting setup for station")

	assetPath := getAssetPath()
	log.Info().Str("asset_path", assetPath).Msg("Using asset path for setup")

	if assetPath == "embedded" {
		log.Info().Msg("Using embedded assets for setup")
		err = dumpAssets("assets", "assets")
		assetPath = "assets"
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to dump assets")
			return
		}
	}

	bookmarks, err := CollectBookmarks(filepath.Join(assetPath, "bookmarks"), stationNum)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to collect bookmarks")
		return
	}
	log.Info().Int("count", len(bookmarks)).Msg("Collected bookmarks")

	err = insertBookmarks("firefox", filepath.Join(assetPath, "etc/firefox/policies/policies.json"), bookmarks)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to insert bookmarks into Firefox policies.json")
		return
	}
	log.Info().Msg("Bookmarks inserted into Firefox policies.json successfully")

	err = insertBookmarks("chromium", filepath.Join(assetPath, "etc/chromium/policies/managed/policies.json"), bookmarks)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to insert bookmarks into Chromium policies.json")
		return
	}
	log.Info().Msg("Bookmarks inserted into Chromium policies.json successfully")

	// Run scripts/post-install.sh
	postInstallPath := filepath.Join(assetPath, "scripts/post-install.sh")
	if err := runScript(postInstallPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to run post-install script")
		return
	}

	// Run scripts/guest-template.sh
	guestTemplatePath := filepath.Join(assetPath, "scripts/guest-template.sh")
	if err := runScript(guestTemplatePath); err != nil {
		log.Fatal().Err(err).Msg("Failed to run guest-template script")
		return
	}

	log.Info().Msg("Setup completed successfully")
}
