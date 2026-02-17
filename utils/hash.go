package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type fileHash struct {
	path string
	hash string
}

func HashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func HashDirectory(root string) (string, error) {

	var hashes []fileHash
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		fh, err := HashFile(path)
		if err != nil {
			return err
		}

		hashes = append(hashes, fileHash{path: relPath, hash: fh})
		return nil
	})

	if err != nil {
		return "", err
	}

	sort.Slice(hashes, func(i, j int) bool {
		return hashes[i].path < hashes[j].path
	})

	combined := sha256.New()
	for _, fh := range hashes {
		combined.Write([]byte(fh.path))
		combined.Write([]byte(fh.hash))
	}

	return hex.EncodeToString(combined.Sum(nil)), nil
}
