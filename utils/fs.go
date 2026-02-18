package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func unzipInto(dest string, zipData []byte) error {
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to read zip data: %w", err)
	}

	for _, entry := range zr.File {
		fullDest := filepath.Join(dest, entry.Name)

		err = os.MkdirAll(filepath.Dir(fullDest), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory for zip entry: %w", err)
		}

		switch entry.FileInfo().IsDir() {
		case true:
			err = os.MkdirAll(fullDest, 0755)
			if err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			log.Debug().Str("dir", entry.Name).Str("path", fullDest).Msg("Created directory for asset")
			continue
		case false:
			rc, err := entry.Open()
			if err != nil {
				return fmt.Errorf("failed to open zip entry: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return fmt.Errorf("failed to read zip entry data: %w", err)
			}

			err = os.WriteFile(fullDest, data, 0644)
			if err != nil {
				return fmt.Errorf("failed to write asset file: %w", err)
			}
			log.Debug().Str("file", entry.Name).Str("path", fullDest).Msg("Asset file written")
			continue
		}
	}
	return nil
}
