package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/MaaXYZ/MaaDebugger/internal/response"
)

type parseInterfaceRequest struct {
	Path string `json:"path"`
}

type rawInterfaceFile struct {
	Name       string             `json:"name"`
	Version    string             `json:"version"`
	Controller []rawInterfaceCtrl `json:"controller"`
	Resource   []rawInterfaceRes  `json:"resource"`
	Task       []rawInterfaceTask `json:"task"`
}

type rawInterfaceCtrl struct {
	Name               string               `json:"name"`
	Type               string               `json:"type"`
	AttachResourcePath []string             `json:"attach_resource_path"`
	Win32              *rawInterfaceWin32   `json:"win32"`
	Gamepad            *rawInterfaceGamepad `json:"gamepad"`
	PlayCover          *rawInterfaceCover   `json:"playcover"`
}

type rawInterfaceWin32 struct {
	ClassRegex  string `json:"class_regex"`
	WindowRegex string `json:"window_regex"`
}

type rawInterfaceGamepad struct {
	ClassRegex  string `json:"class_regex"`
	WindowRegex string `json:"window_regex"`
}

type rawInterfaceCover struct {
	UUID string `json:"uuid"`
}

type rawInterfaceRes struct {
	Name  string   `json:"name"`
	Label string   `json:"label"`
	Path  []string `json:"path"`
}

type rawInterfaceTask struct {
	Name string `json:"name"`
}

type interfaceParseResponse struct {
	InterfacePath        string                    `json:"interface_path"`
	BaseDir              string                    `json:"base_dir"`
	Name                 string                    `json:"name"`
	Version              string                    `json:"version"`
	ControllerCandidates []interfaceControllerItem `json:"controller_candidates"`
	ResourceCandidates   []interfaceResourceItem   `json:"resource_candidates"`
	TaskCandidates       []interfaceTaskItem       `json:"task_candidates"`
}

type interfaceControllerItem struct {
	Name                string   `json:"name"`
	Type                string   `json:"type"`
	ClassRegex          string   `json:"class_regex,omitempty"`
	WindowRegex         string   `json:"window_regex,omitempty"`
	UUID                string   `json:"uuid,omitempty"`
	AttachResourcePaths []string `json:"attach_resource_paths,omitempty"`
}

type interfaceResourceItem struct {
	Name          string                 `json:"name"`
	Label         string                 `json:"label,omitempty"`
	ResolvedPaths []interfaceResolvedRef `json:"resolved_paths"`
}

type interfaceResolvedRef struct {
	Source string `json:"source"`
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
}

type interfaceTaskItem struct {
	Name string `json:"name"`
}

func (r *router) handleInterfaceParse(w http.ResponseWriter, req *http.Request) {
	var payload parseInterfaceRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		response.Fail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	interfacePath := strings.TrimSpace(payload.Path)
	if interfacePath == "" {
		response.Fail(w, http.StatusBadRequest, "path is required")
		return
	}

	result, err := parseInterfaceFile(interfacePath)
	if err != nil {
		log.Warn().Err(err).Str("path", interfacePath).Msg("[Interface] parse failed")
		response.Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Info().
		Str("path", result.InterfacePath).
		Int("controllers", len(result.ControllerCandidates)).
		Int("resources", len(result.ResourceCandidates)).
		Int("tasks", len(result.TaskCandidates)).
		Msg("[Interface] parse succeeded")
	response.OK(w, result)
}

func parseInterfaceFile(interfacePath string) (*interfaceParseResponse, error) {
	absPath, err := filepath.Abs(interfacePath)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var raw rawInterfaceFile
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(absPath)
	result := &interfaceParseResponse{
		InterfacePath:        filepath.Clean(absPath),
		BaseDir:              filepath.Clean(baseDir),
		Name:                 raw.Name,
		Version:              raw.Version,
		ControllerCandidates: make([]interfaceControllerItem, 0, len(raw.Controller)),
		ResourceCandidates:   make([]interfaceResourceItem, 0, len(raw.Resource)),
		TaskCandidates:       make([]interfaceTaskItem, 0, len(raw.Task)),
	}

	for _, item := range raw.Controller {
		candidate := interfaceControllerItem{
			Name:                item.Name,
			Type:                strings.ToLower(strings.TrimSpace(item.Type)),
			AttachResourcePaths: resolvePaths(baseDir, item.AttachResourcePath),
		}
		switch candidate.Type {
		case "win32":
			if item.Win32 != nil {
				candidate.ClassRegex = item.Win32.ClassRegex
				candidate.WindowRegex = item.Win32.WindowRegex
			}
		case "gamepad":
			if item.Gamepad != nil {
				candidate.ClassRegex = item.Gamepad.ClassRegex
				candidate.WindowRegex = item.Gamepad.WindowRegex
			}
		case "playcover":
			if item.PlayCover != nil {
				candidate.UUID = item.PlayCover.UUID
			}
		}
		result.ControllerCandidates = append(result.ControllerCandidates, candidate)
	}

	for _, item := range raw.Resource {
		candidate := interfaceResourceItem{
			Name:          item.Name,
			Label:         item.Label,
			ResolvedPaths: make([]interfaceResolvedRef, 0, len(item.Path)),
		}
		for _, source := range item.Path {
			resolved := resolvePath(baseDir, source)
			_, statErr := os.Stat(resolved)
			candidate.ResolvedPaths = append(candidate.ResolvedPaths, interfaceResolvedRef{
				Source: source,
				Path:   resolved,
				Exists: statErr == nil,
			})
		}
		result.ResourceCandidates = append(result.ResourceCandidates, candidate)
	}

	for _, item := range raw.Task {
		result.TaskCandidates = append(result.TaskCandidates, interfaceTaskItem{Name: item.Name})
	}

	return result, nil
}

func resolvePaths(baseDir string, rawPaths []string) []string {
	if len(rawPaths) == 0 {
		return nil
	}
	result := make([]string, 0, len(rawPaths))
	for _, rawPath := range rawPaths {
		trimmed := strings.TrimSpace(rawPath)
		if trimmed == "" {
			continue
		}
		result = append(result, resolvePath(baseDir, trimmed))
	}
	return result
}

func resolvePath(baseDir, rawPath string) string {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" {
		return ""
	}
	if filepath.IsAbs(trimmed) {
		return filepath.Clean(trimmed)
	}
	return filepath.Clean(filepath.Join(baseDir, trimmed))
}
