package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/inconshreveable/go-update"
)

const githubReleasesURL = "https://api.github.com/repos/saeid-rez/crewup/releases/latest"

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

// CheckForUpdate checks for a newer version and prints a notice if one exists.
// This is non-blocking — errors are silently swallowed.
func CheckForUpdate(currentVersion string) {
	if currentVersion == "dev" {
		return
	}

	latest, err := fetchLatestRelease()
	if err != nil {
		return
	}

	latestVersion := stripV(latest.TagName)
	if latestVersion == currentVersion {
		return
	}

	fmt.Printf("\n┌─────────────────────────────────────────────┐\n")
	fmt.Printf("│  🆕 Update available: %s → %s\n", currentVersion, latestVersion)
	fmt.Printf("│  Run: crewup update                         │\n")
	fmt.Printf("└─────────────────────────────────────────────┘\n\n")
}

// Update downloads and installs the latest release binary.
func Update(currentVersion string) error {
	// Detect Homebrew install
	if isHomebrewInstall() {
		fmt.Println("🍺 You installed crewup via Homebrew.")
		fmt.Println("   Run: brew upgrade crewup")
		return nil
	}

	fmt.Println("🔄 Fetching latest release info...")

	latest, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("could not reach GitHub: %w", err)
	}

	latestVersion := stripV(latest.TagName)
	if latestVersion == currentVersion {
		fmt.Println("✓ Already on the latest version!")
		return nil
	}

	assetURL, checksumURL, assetName := buildDownloadURL(latest.TagName)

	fmt.Printf("📦 Downloading crewup %s...\n", latestVersion)
	archiveData, err := downloadURL(assetURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Verify checksum
	fmt.Println("🔐 Verifying checksum...")
	if err := verifyChecksum(archiveData, checksumURL, assetName); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	// Extract binary from archive
	binary, err := extractBinary(archiveData, assetURL)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Replace current executable
	if err := update.Apply(bytes.NewReader(binary), update.Options{}); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied replacing binary. Try running with sudo or check file permissions")
		}
		return fmt.Errorf("could not replace binary: %w", err)
	}

	fmt.Printf("✓ Updated to %s! Restart crewup to use the new version.\n", latestVersion)
	return nil
}

// isHomebrewInstall returns true if the current binary is in a Homebrew-managed path.
func isHomebrewInstall() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	// Resolve symlinks so that /usr/local/bin/crewup → /opt/homebrew/Cellar/... is detected.
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	lower := strings.ToLower(resolved)
	homebrewPrefixes := []string{
		"/opt/homebrew/",
		"/usr/local/cellar/",
		"/usr/local/homebrew/",
		"/home/linuxbrew/",
		"/homebrew/",
		"/cellar/",
	}
	for _, prefix := range homebrewPrefixes {
		if strings.Contains(lower, prefix) {
			return true
		}
	}
	return false
}

func buildDownloadURL(tag string) (binaryURL, checksumURL, assetName string) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}

	// GoReleaser name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}" — no tag in filename.
	assetName = fmt.Sprintf("crewup_%s_%s%s", goos, goarch, ext)
	base := fmt.Sprintf("https://github.com/saeid-rez/crewup/releases/download/%s/", tag)
	return base + assetName, base + "checksums.txt", assetName
}

func downloadURL(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

func verifyChecksum(data []byte, checksumURL, assetName string) error {
	checksumData, err := downloadURL(checksumURL)
	if err != nil {
		return fmt.Errorf("could not download checksums: %w", err)
	}

	// Parse "sha256  filename" lines
	lines := strings.Split(string(checksumData), "\n")
	var expectedHash string
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == assetName {
			expectedHash = parts[0]
			break
		}
	}
	if expectedHash == "" {
		return fmt.Errorf("checksum not found for %s", assetName)
	}

	sum := sha256.Sum256(data)
	actualHash := hex.EncodeToString(sum[:])
	if actualHash != expectedHash {
		return fmt.Errorf("expected %s, got %s", expectedHash, actualHash)
	}
	return nil
}

func extractBinary(data []byte, assetURL string) ([]byte, error) {
	if strings.HasSuffix(assetURL, ".zip") {
		return extractFromZip(data)
	}
	return extractFromTarGz(data)
}

func extractFromTarGz(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// Find the crewup binary
		name := strings.TrimPrefix(hdr.Name, "./")
		if hdr.Typeflag == tar.TypeReg && (name == "crewup" || name == "crewup.exe") {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("crewup binary not found in archive")
}

func extractFromZip(data []byte) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, f := range r.File {
		name := strings.TrimPrefix(f.Name, "./")
		if name == "crewup" || name == "crewup.exe" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("crewup binary not found in zip archive")
}

func fetchLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: 3 * time.Second}

	req, err := http.NewRequest("GET", githubReleasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func stripV(tag string) string {
	if len(tag) > 0 && tag[0] == 'v' {
		return tag[1:]
	}
	return tag
}
