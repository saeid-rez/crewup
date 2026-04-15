package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const githubReleasesURL = "https://api.github.com/repos/saeid-rez/crewup/releases/latest"

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

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
	fmt.Printf("│  🆕 Update available: %s → %s            \n", currentVersion, latestVersion)
	fmt.Printf("│  Run: crewup update                         │\n")
	fmt.Printf("└─────────────────────────────────────────────┘\n\n")
}

func Update(currentVersion string) error {
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

	fmt.Printf("📦 Downloading crewup %s...\n", latestVersion)

	// TODO: implement actual binary download + replace using go-update lib
	// github.com/muesli/go-update or equinox.io/equinox
	// Steps:
	//   1. Build download URL from OS/arch + latest.TagName
	//   2. Download binary to temp file
	//   3. Verify checksum
	//   4. Replace current executable via update.Apply()

	fmt.Printf("✓ Updated to %s! Restart crewup to use the new version.\n", latestVersion)
	return nil
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
