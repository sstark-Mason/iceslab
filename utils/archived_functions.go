package utils

import (
	"os"

	"github.com/rs/zerolog/log"
)

func CheckForUpdatesByComparingHashes() (bool, bool, error) {
	localManifest, err := GenerateManifest()
	if err != nil {
		return false, false, err
	}

	log.Debug().Str("binary_hash", localManifest.BinaryHash).Msg("Generated current binary hash")
	log.Debug().Str("assets_hash", localManifest.AssetsHash).Msg("Generated current assets hash")

	remoteManifest, err := FetchRemoteManifest()
	if err != nil {
		return false, false, err
	}

	log.Debug().Str("remote_binary_hash", remoteManifest.BinaryHash).Msg("Fetched remote binary hash")
	log.Debug().Str("remote_assets_hash", remoteManifest.AssetsHash).Msg("Fetched remote assets hash")

	binaryUpdateAvailable := localManifest.BinaryHash != remoteManifest.BinaryHash
	assetsUpdateAvailable := localManifest.AssetsHash != remoteManifest.AssetsHash

	return binaryUpdateAvailable, assetsUpdateAvailable, nil
}

func (c *Client) CompareRemoteManifestETag() (bool, error) {

	remoteManifestETag, err := c.FetchManifestETag()
	if err != nil {
		return true, err
	}

	log.Debug().Str("remote_etag", remoteManifestETag).Msg("Fetched remote manifest ETag")

	localManifestETag, err := os.ReadFile(".manifest_etag")
	if err != nil {
		log.Info().Msg("No local manifest ETag found; treating as first run")
		// Save the remote ETag locally for future comparisons
		err = os.WriteFile(".manifest_etag", []byte(remoteManifestETag), 0644)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to save manifest ETag locally")
		} else {
			log.Debug().Str("remote_etag", remoteManifestETag).Msg("Saved remote manifest ETag locally")
		}
		return true, nil
	}

	log.Debug().Str("local_etag", string(localManifestETag)).Msg("Read local manifest ETag")

	if string(localManifestETag) == remoteManifestETag {
		log.Info().Msg("Local manifest ETag matches remote; no updates available")
		return false, nil
	}

	log.Info().Msg("Manifest ETag mismatch; updates available")
	return true, nil
}
