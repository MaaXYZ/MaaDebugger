package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/MaaXYZ/MaaDebugger/internal/buildinfo"
	"github.com/MaaXYZ/MaaDebugger/internal/configstore"
	"github.com/MaaXYZ/MaaDebugger/internal/console"
)

const (
	baseURL    = "https://api.maafw.com/MaaDebugger"
	latestURL  = baseURL + "/latest.json"
	nightlyURL = baseURL + "/nightly.json"

	storeKeyLastCheck = "__updater_last_check"

	nightlyCooldown = 1 * time.Hour
	latestCooldown  = 12 * time.Hour
)

// VersionEntry represents a single version entry (used in both latest and nightly).
type VersionEntry struct {
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Note      string `json:"note,omitempty"`
}

// LatestResponse is the response format from latest.json.
type LatestResponse struct {
	Release    VersionEntry `json:"release"`
	PreRelease VersionEntry `json:"preRelease"`
}

// NightlyResponse is the response format from nightly.json.
type NightlyResponse struct {
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

// CheckResult is returned to the frontend.
type CheckResult struct {
	HasUpdate      bool   `json:"has_update"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Note           string `json:"note,omitempty"`
	Nightly        bool   `json:"nightly"`
}

func IsCommitHash(s string) bool {
	return len(s) == 40
}

// CheckUpdate fetches the latest version info and compares with current version.
// If nightly is true, it fetches from the nightly endpoint instead of latest.
// Returns nil result when version is "dev" (skips check).
func CheckUpdate(nightly bool) (*CheckResult, error) {
	// Skip check for dev builds
	if buildinfo.Version == "dev" {
		return &CheckResult{
			HasUpdate:      false,
			CurrentVersion: "dev",
			LatestVersion:  "",
			Nightly:        nightly,
		}, nil
	}

	apiURL := latestURL
	if nightly {
		apiURL = nightlyURL
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch update info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	currentVersion := buildinfo.Version
	var latestVersion string
	var note string

	if nightly {
		var nightlyResp NightlyResponse
		if err := json.Unmarshal(body, &nightlyResp); err != nil {
			return nil, fmt.Errorf("failed to parse nightly info: %w", err)
		}
		latestVersion = nightlyResp.Version
	} else {
		var latestResp LatestResponse
		if err := json.Unmarshal(body, &latestResp); err != nil {
			return nil, fmt.Errorf("failed to parse latest info: %w", err)
		}
		// Prefer release, fall back to preRelease
		if latestResp.Release.Version != "" {
			latestVersion = latestResp.Release.Version
			note = latestResp.Release.Note
		} else if latestResp.PreRelease.Version != "" {
			latestVersion = latestResp.PreRelease.Version
			note = latestResp.PreRelease.Note
		}
	}

	hasUpdate := latestVersion != "" && latestVersion != currentVersion

	result := &CheckResult{
		HasUpdate:      hasUpdate,
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		Note:           note,
		Nightly:        nightly,
	}

	// Log to console when an update is available
	if hasUpdate {
		channel := "latest"
		if nightly {
			channel = "nightly"
		}
		log.Info().
			Str("current", currentVersion).
			Str("latest", latestVersion).
			Str("channel", channel).
			Msg("New version available!")

		console.Warnf("New version available! (%s)", channel)
		console.Infof("  Current: %s%s%s", console.Red, currentVersion, console.Reset)
		console.Infof("  Latest:  %s%s%s", console.BrightGreen, latestVersion, console.Reset)
		if note != "" {
			console.Infof("  Note:    %s%s%s", console.BrightCyan, note, console.Reset)
		}
	}

	return result, nil
}

// AutoCheckUpdate performs a check with cooldown enforcement via configstore.
// nightly: 1h cooldown, non-nightly: 12h cooldown.
// Returns nil, nil if the check is skipped due to cooldown.
func AutoCheckUpdate(store *configstore.Store, nightly bool) (*CheckResult, error) {
	cooldown := latestCooldown
	if nightly {
		cooldown = nightlyCooldown
	}

	// Read last check timestamp from store
	if v, ok := store.Get(storeKeyLastCheck); ok {
		if ts, ok := v.(float64); ok {
			lastCheck := time.Unix(int64(ts), 0)
			if time.Since(lastCheck) < cooldown {
				log.Debug().
					Time("lastCheck", lastCheck).
					Dur("cooldown", cooldown).
					Msg("skipping auto update check (cooldown)")
				return nil, nil
			}
		}
	}

	result, err := CheckUpdate(nightly)
	if err != nil {
		return nil, err
	}

	// Update last check timestamp
	store.Set(storeKeyLastCheck, time.Now().Unix())

	return result, nil
}
