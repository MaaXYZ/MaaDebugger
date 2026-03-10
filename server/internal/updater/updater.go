package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"

	"github.com/MaaXYZ/MaaDebugger/internal/buildinfo"
	"github.com/MaaXYZ/MaaDebugger/internal/configstore"
	"github.com/MaaXYZ/MaaDebugger/internal/console"
)

const (
	baseURL    = "https://api.maafw.com/MaaDebugger"
	latestURL  = baseURL + "/latest.json"
	nightlyURL = baseURL + "/nightly.json"

	storeKeySettings       = "updateSettings"
	storeKeyLastCheck      = "lastCheck"
	storeKeyShowPreRelease = "showPreRelease"

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
	Track          string `json:"track"`
}

type CheckOptions struct {
	Nightly           bool
	IncludePreRelease bool
}

func IsCommitHash(s string) bool {
	return len(s) == 40
}

func IsNightlyBuild(channel string, version string) bool {
	return (channel == "npm" || channel == "github") && IsCommitHash(version)
}

func LoadIncludePreRelease(store *configstore.Store) bool {
	if store == nil {
		return false
	}

	v, ok := store.Get(storeKeySettings)
	if !ok {
		return false
	}

	settings, ok := v.(map[string]any)
	if !ok {
		return false
	}

	showPre, ok := settings[storeKeyShowPreRelease]
	if !ok {
		return false
	}

	return asBool(showPre)
}

func asBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		parsed, err := strconv.ParseBool(v)
		return err == nil && parsed
	case float64:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	default:
		return false
	}
}

func parseBuildTimestamp(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty build timestamp")
	}

	if unix, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		if len(trimmed) >= 13 {
			return time.UnixMilli(unix), nil
		}
		return time.Unix(unix, 0), nil
	}

	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid build timestamp %q", value)
	}

	return parsed, nil
}

func normalizeSemver(version string) string {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "v") {
		return trimmed
	}
	return "v" + trimmed
}

func compareSemver(currentVersion string, candidateVersion string) (int, error) {
	current := normalizeSemver(currentVersion)
	candidate := normalizeSemver(candidateVersion)

	if !semver.IsValid(current) {
		return 0, fmt.Errorf("invalid current semver: %s", currentVersion)
	}
	if !semver.IsValid(candidate) {
		return 0, fmt.Errorf("invalid candidate semver: %s", candidateVersion)
	}

	return semver.Compare(candidate, current), nil
}

func pickLatestVersion(currentVersion string, latestResp LatestResponse, includePreRelease bool) (string, string, string, error) {
	type candidate struct {
		track   string
		version string
		note    string
	}

	candidates := make([]candidate, 0, 2)
	if latestResp.Release.Version != "" {
		candidates = append(candidates, candidate{
			track:   "release",
			version: latestResp.Release.Version,
			note:    latestResp.Release.Note,
		})
	}
	if includePreRelease && latestResp.PreRelease.Version != "" {
		candidates = append(candidates, candidate{
			track:   "pre-release",
			version: latestResp.PreRelease.Version,
			note:    latestResp.PreRelease.Note,
		})
	}

	bestTrack := "release"
	bestVersion := ""
	bestNote := ""
	bestCmp := 0

	for _, candidate := range candidates {
		cmp, err := compareSemver(currentVersion, candidate.version)
		if err != nil {
			return "", "", "", err
		}
		if bestVersion == "" || cmp > bestCmp {
			bestTrack = candidate.track
			bestVersion = candidate.version
			bestNote = candidate.note
			bestCmp = cmp
		}
	}

	return bestVersion, bestNote, bestTrack, nil
}

// CheckUpdate fetches the latest version info and compares with current version.
// If nightly is true, it fetches from the nightly endpoint instead of latest.
// Returns nil result when version is "dev" (skips check).
func CheckUpdate(opts CheckOptions) (*CheckResult, error) {
	// Skip check for dev builds
	if buildinfo.Version == "dev" {
		return &CheckResult{
			HasUpdate:      false,
			CurrentVersion: "dev",
			LatestVersion:  "",
			Nightly:        opts.Nightly,
			Track:          "dev",
		}, nil
	}

	apiURL := latestURL
	if opts.Nightly {
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
	track := "release"
	hasUpdate := false

	if opts.Nightly {
		var nightlyResp NightlyResponse
		if err := json.Unmarshal(body, &nightlyResp); err != nil {
			return nil, fmt.Errorf("failed to parse nightly info: %w", err)
		}
		latestVersion = nightlyResp.Version
		track = "nightly"

		currentBuildTime, err := parseBuildTimestamp(buildinfo.BuildTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse current build time: %w", err)
		}
		latestBuildTime, err := parseBuildTimestamp(nightlyResp.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse nightly build time: %w", err)
		}
		hasUpdate = latestBuildTime.After(currentBuildTime)
	} else {
		var latestResp LatestResponse
		if err := json.Unmarshal(body, &latestResp); err != nil {
			return nil, fmt.Errorf("failed to parse latest info: %w", err)
		}

		latestVersion, note, track, err = pickLatestVersion(currentVersion, latestResp, opts.IncludePreRelease)
		if err != nil {
			return nil, fmt.Errorf("failed to compare release versions: %w", err)
		}

		if latestVersion != "" {
			cmp, err := compareSemver(currentVersion, latestVersion)
			if err != nil {
				return nil, fmt.Errorf("failed to compare selected version: %w", err)
			}
			hasUpdate = cmp > 0
		}
	}

	result := &CheckResult{
		HasUpdate:      hasUpdate,
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		Note:           note,
		Nightly:        opts.Nightly,
		Track:          track,
	}

	// Log to console when an update is available
	if hasUpdate {
		log.Info().
			Str("current", currentVersion).
			Str("latest", latestVersion).
			Str("channel", track).
			Msg("New version available!")

		console.Warnf("New version available! (%s)", track)
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
func AutoCheckUpdate(store *configstore.Store, opts CheckOptions) (*CheckResult, error) {
	cooldown := latestCooldown
	if opts.Nightly {
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

	result, err := CheckUpdate(opts)
	if err != nil {
		return nil, err
	}

	// Update last check timestamp
	store.Set(storeKeyLastCheck, time.Now().Unix())

	return result, nil
}
