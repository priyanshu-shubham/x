package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// GetLatestRelease fetches the latest release info from GitHub
func GetLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GitHubOwner, GitHubRepo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// GetAssetName returns the expected asset name for the current platform
func GetAssetName() string {
	ext := "tar.gz"
	if runtime.GOOS == OSWindows {
		ext = "zip"
	}
	return fmt.Sprintf("x_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)
}

// DownloadFile downloads a file from a URL
func DownloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// ExtractBinary extracts the binary from the downloaded archive
func ExtractBinary(archivePath, destDir string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, destDir)
	}
	return extractTarGz(archivePath, destDir)
}

func extractTarGz(archivePath, destDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	binaryName := "x"

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Typeflag == tar.TypeReg && (header.Name == "x" || header.Name == "x.exe") {
			binaryName = header.Name
			destPath := filepath.Join(destDir, binaryName)
			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", err
			}
			outFile.Close()
			return destPath, nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

func extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "x" || f.Name == "x.exe" {
			destPath := filepath.Join(destDir, f.Name)

			rc, err := f.Open()
			if err != nil {
				return "", err
			}

			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
			if err != nil {
				rc.Close()
				return "", err
			}

			_, err = io.Copy(outFile, rc)
			rc.Close()
			outFile.Close()
			if err != nil {
				return "", err
			}

			return destPath, nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// RunUpgrade performs the upgrade
func RunUpgrade() error {
	fmt.Printf("Current version: %s\n", Version)
	fmt.Println("Checking for updates...")

	release, err := GetLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := release.TagName
	fmt.Printf("Latest version: %s\n", latestVersion)

	// Compare versions (strip 'v' prefix for comparison)
	currentClean := strings.TrimPrefix(Version, "v")
	latestClean := strings.TrimPrefix(latestVersion, "v")

	if currentClean == latestClean {
		fmt.Println("Already up to date!")
		return nil
	}

	// Find the right asset
	assetName := GetAssetName()
	var downloadURL string

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no release found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Downloading %s...\n", assetName)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "x-upgrade-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Download
	archivePath := filepath.Join(tempDir, assetName)
	if err := DownloadFile(downloadURL, archivePath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Extract
	binaryPath, err := ExtractBinary(archivePath, tempDir)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine current executable: %w", err)
	}
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %w", err)
	}

	// Replace the binary
	// On Linux and Windows, we can't overwrite a running executable directly.
	// Remove/rename the old binary first, then place the new one at the same path.
	oldPath := currentExe + ".old"
	os.Remove(oldPath) // Remove any existing .old file
	if err := os.Rename(currentExe, oldPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Copy new binary
	newBinary, err := os.ReadFile(binaryPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(currentExe, newBinary, 0755); err != nil {
		// Try to restore the old binary on failure
		os.Rename(oldPath, currentExe)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Clean up the old binary (best-effort)
	os.Remove(oldPath)

	fmt.Printf("Upgraded to %s\n", latestVersion)
	return nil
}
