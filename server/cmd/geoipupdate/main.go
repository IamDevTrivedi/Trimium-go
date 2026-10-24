package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	editionID = "GeoLite2-City"

	// maxMMDBFileSize bounds how much data is extracted from the archive so a
	// malicious or corrupted download cannot exhaust disk space.
	maxMMDBFileSize = int64(512 << 20) // 512 MiB
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fail: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Success: GeoLite2-City database updated.")
}

func run() error {
	key := os.Getenv("MAXMIND_LICENSE_KEY")
	if key == "" {
		return fmt.Errorf("MAXMIND_LICENSE_KEY environment variable is not set")
	}

	targetFile, err := targetPath()
	if err != nil {
		return err
	}

	q := url.Values{}
	q.Set("edition_id", editionID)
	q.Set("license_key", key)
	q.Set("suffix", "tar.gz")
	downloadURL := "https://download.maxmind.com/app/geoip_download?" + q.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("build download request: %w", sanitize(err, key))
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", sanitize(err, key))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("gunzip: %w", err)
	}
	defer gz.Close()

	dir := filepath.Dir(targetFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Write to a temp file in the same dir, then atomically rename,
	// so a running app never sees a half-written DB.
	tmp, err := os.CreateTemp(dir, ".geoip-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after successful rename

	found := false
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			tmp.Close()
			return fmt.Errorf("tar: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && strings.HasSuffix(hdr.Name, ".mmdb") {
			written, err := io.Copy(tmp, io.LimitReader(tr, maxMMDBFileSize+1))
			if err != nil {
				tmp.Close()
				return fmt.Errorf("extract MMDB file: %w", err)
			}
			if written > maxMMDBFileSize {
				tmp.Close()
				return fmt.Errorf("extracted MMDB file exceeds the %d byte limit", maxMMDBFileSize)
			}
			found = true
			break
		}
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("MMDB file not found in archive")
	}

	return os.Rename(tmpName, targetFile)
}

// targetPath returns the operator-configured destination for the database.
func targetPath() (string, error) {
	raw := os.Getenv("GEOIP_DB_PATH")
	if raw == "" {
		return "", fmt.Errorf("GEOIP_DB_PATH environment variable is not set")
	}

	cleaned := filepath.Clean(raw)
	if cleaned == "." || cleaned == string(filepath.Separator) {
		return "", fmt.Errorf("invalid GEOIP_DB_PATH %q: must point to a file", raw)
	}

	return cleaned, nil
}

// sanitize strips the license key from URL errors so it never hits logs.
func sanitize(err error, key string) error {
	return fmt.Errorf("%s", strings.ReplaceAll(err.Error(), key, "***"))
}
