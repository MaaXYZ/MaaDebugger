package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	Import     []string           `json:"import"`
	Languages  map[string]string  `json:"languages"`
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
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Entry       string   `json:"entry"`
	Description any      `json:"description"`
	Controller  []string `json:"controller"`
	Resource    []string `json:"resource"`
	Option      []string `json:"option"`
}

type rawImportedTaskFile struct {
	Task   []rawInterfaceTask               `json:"task"`
	Option map[string]rawImportedTaskOption `json:"option"`
	Import []string                         `json:"import"`
}

type rawImportedTaskOption struct {
	Type        string                  `json:"type"`
	Label       string                  `json:"label"`
	Description any                     `json:"description"`
	DefaultCase string                  `json:"default_case"`
	Cases       []rawImportedOptionCase `json:"cases"`
}

type rawImportedOptionCase struct {
	Name             string         `json:"name"`
	Label            string         `json:"label"`
	Description      any            `json:"description"`
	PipelineOverride map[string]any `json:"pipeline_override"`
}

type interfaceParseResponse struct {
	InterfacePath        string                              `json:"interface_path"`
	BaseDir              string                              `json:"base_dir"`
	Name                 string                              `json:"name"`
	Version              string                              `json:"version"`
	Languages            map[string]string                   `json:"languages,omitempty"`
	LocaleValues         map[string]map[string]string        `json:"locale_values,omitempty"`
	Imports              []interfaceResolvedRef              `json:"imports,omitempty"`
	ControllerCandidates []interfaceControllerItem           `json:"controller_candidates"`
	ResourceCandidates   []interfaceResourceItem             `json:"resource_candidates"`
	TaskCandidates       []interfaceTaskItem                 `json:"task_candidates"`
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
	Name            string                    `json:"name"`
	Label           string                    `json:"label,omitempty"`
	Entry           string                    `json:"entry,omitempty"`
	Description     string                    `json:"description,omitempty"`
	Controllers     []string                  `json:"controllers,omitempty"`
	Resources       []string                  `json:"resources,omitempty"`
	Options         []string                  `json:"options,omitempty"`
	OptionDefs      []interfaceTaskOptionItem `json:"option_defs,omitempty"`
	Source          string                    `json:"source,omitempty"`
	SourceInterface string                    `json:"source_interface,omitempty"`
}

type interfaceTaskOptionItem struct {
	Name         string                    `json:"name"`
	Type         string                    `json:"type,omitempty"`
	Label        string                    `json:"label,omitempty"`
	Description  string                    `json:"description,omitempty"`
	DefaultCase  string                    `json:"default_case,omitempty"`
	Cases        []interfaceTaskOptionCase `json:"cases,omitempty"`
	Source       string                    `json:"source,omitempty"`
	ResolvedFrom string                    `json:"resolved_from,omitempty"`
}

type interfaceTaskOptionCase struct {
	Name                 string         `json:"name"`
	Label                string         `json:"label,omitempty"`
	Description          string         `json:"description,omitempty"`
	PipelineOverrideKeys []string       `json:"pipeline_override_keys,omitempty"`
	PipelineOverride     map[string]any `json:"pipeline_override,omitempty"`
}

type parsedImportedTasks struct {
	Tasks []interfaceTaskItem
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
		Int("imports", len(result.Imports)).
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
	if err := unmarshalLenientJSON(content, &raw); err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(absPath)
	result := &interfaceParseResponse{
		InterfacePath:        filepath.Clean(absPath),
		BaseDir:              filepath.Clean(baseDir),
		Name:                 raw.Name,
		Version:              raw.Version,
		Languages:            compactStringMap(raw.Languages),
		LocaleValues:         map[string]map[string]string{},
		Imports:              make([]interfaceResolvedRef, 0, len(raw.Import)),
		ControllerCandidates: make([]interfaceControllerItem, 0, len(raw.Controller)),
		ResourceCandidates:   make([]interfaceResourceItem, 0, len(raw.Resource)),
		TaskCandidates:       make([]interfaceTaskItem, 0, len(raw.Task)),
	}
	localeValues, err := parseInterfaceLocales(baseDir, result.Languages)
	if err != nil {
		return nil, err
	}
	result.LocaleValues = localeValues

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
		result.TaskCandidates = append(result.TaskCandidates, buildTaskCandidate(item, result.InterfacePath, nil))
	}

	imported, imports, err := parseImportedTasks(baseDir, raw.Import, result.InterfacePath)
	if err != nil {
		return nil, err
	}
	result.Imports = append(result.Imports, imports...)
	result.TaskCandidates = append(result.TaskCandidates, imported.Tasks...)

	return result, nil
}

func parseImportedTasks(baseDir string, imports []string, sourceInterface string) (*parsedImportedTasks, []interfaceResolvedRef, error) {
	resolvedImports := make([]interfaceResolvedRef, 0, len(imports))
	aggregated := &parsedImportedTasks{Tasks: make([]interfaceTaskItem, 0)}
	visited := map[string]bool{}

	var walk func(currentBaseDir string, importPaths []string) error
	walk = func(currentBaseDir string, importPaths []string) error {
		for _, importPath := range importPaths {
			trimmed := strings.TrimSpace(importPath)
			if trimmed == "" {
				continue
			}

			resolved := resolvePath(currentBaseDir, trimmed)
			_, statErr := os.Stat(resolved)
			resolvedImports = append(resolvedImports, interfaceResolvedRef{
				Source: trimmed,
				Path:   resolved,
				Exists: statErr == nil,
			})
			if statErr != nil {
				return fmt.Errorf("import file not found: %s", resolved)
			}
			if visited[resolved] {
				continue
			}
			visited[resolved] = true

			content, err := os.ReadFile(resolved)
			if err != nil {
				return err
			}

			var imported rawImportedTaskFile
			if err := unmarshalLenientJSON(content, &imported); err != nil {
				return fmt.Errorf("parse import %s: %w", resolved, err)
			}

			for _, task := range imported.Task {
				aggregated.Tasks = append(aggregated.Tasks, buildTaskCandidate(task, sourceInterface, imported.Option))
			}

			if len(imported.Import) > 0 {
				if err := walk(filepath.Dir(resolved), imported.Import); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := walk(baseDir, imports); err != nil {
		return nil, nil, err
	}

	return aggregated, resolvedImports, nil
}

func buildTaskCandidate(raw rawInterfaceTask, sourceInterface string, optionDefs map[string]rawImportedTaskOption) interfaceTaskItem {
	candidate := interfaceTaskItem{
		Name:            raw.Name,
		Label:           raw.Label,
		Entry:           raw.Entry,
		Description:     stringifyText(raw.Description),
		Controllers:     compactStrings(raw.Controller),
		Resources:       compactStrings(raw.Resource),
		Options:         compactStrings(raw.Option),
		OptionDefs:      make([]interfaceTaskOptionItem, 0, len(raw.Option)),
		Source:          sourceInterface,
		SourceInterface: sourceInterface,
	}

	for _, optionName := range candidate.Options {
		def, ok := optionDefs[optionName]
		if !ok {
			candidate.OptionDefs = append(candidate.OptionDefs, interfaceTaskOptionItem{Name: optionName})
			continue
		}
		candidate.OptionDefs = append(candidate.OptionDefs, buildTaskOptionCandidate(optionName, def, sourceInterface))
	}

	return candidate
}

func buildTaskOptionCandidate(name string, raw rawImportedTaskOption, sourceInterface string) interfaceTaskOptionItem {
	item := interfaceTaskOptionItem{
		Name:         name,
		Type:         strings.TrimSpace(raw.Type),
		Label:        raw.Label,
		Description:  stringifyText(raw.Description),
		DefaultCase:  raw.DefaultCase,
		Cases:        make([]interfaceTaskOptionCase, 0, len(raw.Cases)),
		Source:       sourceInterface,
		ResolvedFrom: sourceInterface,
	}

	for _, rawCase := range raw.Cases {
		keys := make([]string, 0, len(rawCase.PipelineOverride))
		for key := range rawCase.PipelineOverride {
			keys = append(keys, key)
		}
		item.Cases = append(item.Cases, interfaceTaskOptionCase{
			Name:                 rawCase.Name,
			Label:                rawCase.Label,
			Description:          stringifyText(rawCase.Description),
			PipelineOverrideKeys: keys,
			PipelineOverride:     rawCase.PipelineOverride,
		})
	}

	return item
}

func compactStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func stringifyText(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			text := stringifyText(item)
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func compactStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		trimmedKey := strings.TrimSpace(key)
		trimmedValue := strings.TrimSpace(value)
		if trimmedKey == "" || trimmedValue == "" {
			continue
		}
		result[trimmedKey] = trimmedValue
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func parseInterfaceLocales(baseDir string, languages map[string]string) (map[string]map[string]string, error) {
	if len(languages) == 0 {
		return nil, nil
	}

	result := make(map[string]map[string]string, len(languages))
	for language, relativePath := range languages {
		resolved := resolvePath(baseDir, relativePath)
		content, err := os.ReadFile(resolved)
		if err != nil {
			return nil, fmt.Errorf("read locale %s: %w", resolved, err)
		}

		var raw any
		if err := unmarshalLenientJSON(content, &raw); err != nil {
			return nil, fmt.Errorf("parse locale %s: %w", resolved, err)
		}

		flattened := map[string]string{}
		flattenLocaleMap("", raw, flattened)
		result[language] = flattened
	}

	return result, nil
}

func flattenLocaleMap(prefix string, value any, out map[string]string) {
	switch v := value.(type) {
	case map[string]any:
		for key, nested := range v {
			nextPrefix := key
			if prefix != "" {
				nextPrefix = prefix + "." + key
			}
			flattenLocaleMap(nextPrefix, nested, out)
		}
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			text := stringifyText(item)
			if text != "" {
				parts = append(parts, text)
			}
		}
		if prefix != "" && len(parts) > 0 {
			out[prefix] = strings.Join(parts, "\n")
		}
	case string:
		if prefix != "" {
			out[prefix] = strings.TrimSpace(v)
		}
	default:
		if prefix != "" {
			text := stringifyText(v)
			if text != "" {
				out[prefix] = text
			}
		}
	}
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

func unmarshalLenientJSON(content []byte, target any) error {
	normalized := stripJSONComments(string(content))
	normalized = stripTrailingCommas(normalized)
	return json.Unmarshal([]byte(normalized), target)
}

func stripJSONComments(input string) string {
	var out strings.Builder
	out.Grow(len(input))

	inString := false
	escaped := false
	inLineComment := false
	inBlockComment := false

	for i := 0; i < len(input); i++ {
		ch := input[i]
		next := byte(0)
		if i+1 < len(input) {
			next = input[i+1]
		}

		switch {
		case inLineComment:
			if ch == '\n' || ch == '\r' {
				inLineComment = false
				out.WriteByte(ch)
			}
			continue
		case inBlockComment:
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		case inString:
			out.WriteByte(ch)
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		default:
			if ch == '"' {
				inString = true
				out.WriteByte(ch)
				continue
			}
			if ch == '/' && next == '/' {
				inLineComment = true
				i++
				continue
			}
			if ch == '/' && next == '*' {
				inBlockComment = true
				i++
				continue
			}
			out.WriteByte(ch)
		}
	}

	return out.String()
}

func stripTrailingCommas(input string) string {
	var out strings.Builder
	out.Grow(len(input))

	inString := false
	escaped := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if inString {
			out.WriteByte(ch)
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			out.WriteByte(ch)
			continue
		}

		if ch == ',' {
			j := i + 1
			for j < len(input) {
				r := rune(input[j])
				if !strconv.IsPrint(r) && r != '\n' && r != '\r' && r != '\t' {
					break
				}
				if !strings.ContainsRune(" \t\r\n", r) {
					break
				}
				j++
			}
			if j < len(input) && (input[j] == '}' || input[j] == ']') {
				continue
			}
		}

		out.WriteByte(ch)
	}

	return out.String()
}
