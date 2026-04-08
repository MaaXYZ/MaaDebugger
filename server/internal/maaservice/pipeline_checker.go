package maaservice

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/tailscale/hujson"
)

type CheckResponse struct {
	Level string `json:"level"`
	Code  string `json:"code"`
	Msg   string `json:"msg"`
	Task  string `json:"task,omitempty"`
	Path  string `json:"path"`
	Line  string `json:"line"`
}

func nodeLevelByCode(code string) string {
	switch code {
	case "mpe-config", "image-path-back-slash", "image-path-dot-slash", "image-path-missing-png", "dynamic-image":
		return "warning"
	default:
		return "error"
	}
}

func newDiag(code string, msg string, path string, line string, task string) CheckResponse {
	return CheckResponse{
		Level: nodeLevelByCode(strings.TrimSpace(code)),
		Code:  strings.TrimSpace(code),
		Msg:   msg,
		Task:  strings.TrimSpace(task),
		Path:  path,
		Line:  line,
	}
}

type PipelineChecker struct {
	mux      sync.Mutex
	cond     *sync.Cond
	paths    []string
	cache    []CheckResponse
	running  bool
	checking bool
	pending  bool
}

func NewPipelineChecker() *PipelineChecker {
	s := &PipelineChecker{}
	s.cond = sync.NewCond(&s.mux)
	return s
}

func (s *PipelineChecker) SetPaths(paths []string) {
	s.mux.Lock()
	s.paths = paths
	s.mux.Unlock()
}

func (s *PipelineChecker) Start() {
	s.mux.Lock()
	s.running = true
	s.mux.Unlock()
}

func (s *PipelineChecker) Stop() {
	s.mux.Lock()
	s.running = false
	s.mux.Unlock()
}

func (s *PipelineChecker) OnPathChanged(_ string) {
	s.mux.Lock()
	running := s.running
	s.mux.Unlock()
	if !running {
		return
	}
	s.triggerAsync(false)
}

func (s *PipelineChecker) GetResult() []CheckResponse {
	s.mux.Lock()
	for s.checking {
		s.cond.Wait()
	}
	result := append([]CheckResponse(nil), s.cache...)
	s.mux.Unlock()
	return result
}

func (s *PipelineChecker) CheckOnce() []CheckResponse {
	paths := s.prepareRun()
	result := runPipelineCheck(paths)
	s.finishRun(result)
	return result
}

func (s *PipelineChecker) triggerAsync(force bool) {
	s.mux.Lock()
	if s.checking {
		s.pending = true
		s.mux.Unlock()
		return
	}
	s.checking = true
	paths := append([]string(nil), s.paths...)
	s.mux.Unlock()

	go func() {
		result := runPipelineCheck(paths)
		s.finishRun(result)

		s.mux.Lock()
		runAgain := s.pending
		s.pending = false
		s.mux.Unlock()
		if runAgain || force {
			s.triggerAsync(false)
		}
	}()
}

func (s *PipelineChecker) prepareRun() []string {
	s.mux.Lock()
	for s.checking {
		s.cond.Wait()
	}
	s.checking = true
	paths := append([]string(nil), s.paths...)
	s.mux.Unlock()
	return paths
}

func (s *PipelineChecker) finishRun(result []CheckResponse) {
	s.mux.Lock()
	s.cache = append([]CheckResponse(nil), result...)
	s.checking = false
	s.cond.Broadcast()
	s.mux.Unlock()
}

type pipelineFile struct {
	path       string
	root       string
	raw        string
	lineStarts []int
	tasks      map[string]map[string]any
}

type taskDecl struct {
	path string
	line string
}

func runPipelineCheck(roots []string) []CheckResponse {
	files := collectPipelineFiles(roots)
	images := collectImageFiles(roots)

	result := make([]CheckResponse, 0, 64)
	parsed := make([]pipelineFile, 0, len(files))
	decls := make(map[string]taskDecl)
	allTasks := make(map[string]struct{})

	for _, file := range files {
		pf, diags := parsePipelineFile(file, roots)
		if len(diags) > 0 {
			result = append(result, diags...)
		}
		if pf == nil {
			continue
		}
		parsed = append(parsed, *pf)
		for name := range pf.tasks {
			if prev, ok := decls[name]; ok {
				line := findTaskLine(pf.raw, pf.lineStarts, name)
				result = append(result, newDiag("conflict-task", fmt.Sprintf("Conflict task %s, previous defined in %s:%s", name, prev.path, prev.line), pf.path, line, name))
				continue
			}
			line := findTaskLine(pf.raw, pf.lineStarts, name)
			decls[name] = taskDecl{path: pf.path, line: line}
			allTasks[name] = struct{}{}
		}
	}

	for i := range parsed {
		result = append(result, checkTaskRules(parsed[i], allTasks, images)...)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Path != result[j].Path {
			return result[i].Path < result[j].Path
		}
		li, ci := splitLineCol(result[i].Line)
		lj, cj := splitLineCol(result[j].Line)
		if li != lj {
			return li < lj
		}
		if ci != cj {
			return ci < cj
		}
		if result[i].Level != result[j].Level {
			return result[i].Level < result[j].Level
		}
		return result[i].Msg < result[j].Msg
	})

	return result
}

func parsePipelineFile(filePath string, roots []string) (*pipelineFile, []CheckResponse) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, []CheckResponse{newDiag("syntax-error", err.Error(), filePath, "1:1", "")}
	}

	raw := string(content)
	lineStarts := buildLineStarts(raw)
	if _, err := hujson.Parse(content); err != nil {
		line, col := parseHujsonLineCol(err)
		return nil, []CheckResponse{newDiag("syntax-error", err.Error(), filePath, fmt.Sprintf("%d:%d", line, col), "")}
	}

	std, err := hujson.Standardize(content)
	if err != nil {
		line, col := parseHujsonLineCol(err)
		return nil, []CheckResponse{newDiag("syntax-error", err.Error(), filePath, fmt.Sprintf("%d:%d", line, col), "")}
	}

	var root map[string]any
	if err := json.Unmarshal(std, &root); err != nil {
		line, col := parseStdJSONLineCol(err)
		return nil, []CheckResponse{newDiag("syntax-error", err.Error(), filePath, fmt.Sprintf("%d:%d", line, col), "")}
	}

	tasks := make(map[string]map[string]any)
	diags := make([]CheckResponse, 0, 2)
	for k, v := range root {
		if strings.HasPrefix(k, "$__mpe") {
			diags = append(diags, newDiag(
				"mpe-config",
				"MPE config detected",
				filePath,
				findTaskLine(raw, lineStarts, k),
				k,
			))
			continue
		}
		if strings.HasPrefix(k, "$") {
			continue
		}
		obj, ok := v.(map[string]any)
		if !ok {
			continue
		}
		tasks[k] = obj
	}

	return &pipelineFile{
		path:       filePath,
		root:       resolveOwnerRoot(filePath, roots),
		raw:        raw,
		lineStarts: lineStarts,
		tasks:      tasks,
	}, diags
}

func checkTaskRules(pf pipelineFile, allTasks map[string]struct{}, images map[string]map[string]struct{}) []CheckResponse {
	anchors := collectAnchors(pf.tasks)
	diags := make([]CheckResponse, 0, 24)
	for taskName, taskObj := range pf.tasks {
		diags = append(diags, checkUnknownTaskRefs(pf, taskName, taskObj, allTasks, anchors)...)
		diags = append(diags, checkUnknownAnchorRefs(pf, taskName, taskObj, anchors)...)
		diags = append(diags, checkTemplateWarningsAndUnknownImage(pf, taskName, taskObj, images)...)
		diags = append(diags, checkUnknownAttr(pf, taskName, taskObj)...)
		diags = append(diags, checkDuplicateNext(pf, taskName, taskObj)...)
	}
	return diags
}

func collectAnchors(tasks map[string]map[string]any) map[string]struct{} {
	anchors := make(map[string]struct{})
	for _, taskObj := range tasks {
		if anchorDecl, ok := taskObj["anchor"]; ok {
			switch v := anchorDecl.(type) {
			case string:
				a := strings.TrimSpace(v)
				if a != "" {
					anchors[a] = struct{}{}
				}
			case map[string]any:
				for name := range v {
					a := strings.TrimSpace(name)
					if a != "" {
						anchors[a] = struct{}{}
					}
				}
			}
		}
		if customAnchorDecl, ok := taskObj["custom_anchor"]; ok {
			for _, anchor := range extractStringValues(customAnchorDecl) {
				a := strings.TrimSpace(anchor)
				if a != "" {
					anchors[a] = struct{}{}
				}
			}
		}
	}
	return anchors
}

func checkUnknownAnchorRefs(pf pipelineFile, taskName string, taskObj map[string]any, anchors map[string]struct{}) []CheckResponse {
	if len(anchors) == 0 {
		return nil
	}
	diags := make([]CheckResponse, 0)
	for _, ref := range extractStringByKey(taskObj, "Anchor") {
		a := strings.TrimSpace(ref)
		if a == "" {
			continue
		}
		if _, ok := anchors[a]; ok {
			continue
		}
		line := findTaskKeyValueLine(pf, taskName, "Anchor", a)
		diags = append(diags, newDiag("unknown-anchor", fmt.Sprintf("Unknown anchor %s", a), pf.path, line, taskName))
	}
	return diags
}

func checkUnknownTaskRefs(pf pipelineFile, taskName string, taskObj map[string]any, allTasks map[string]struct{}, anchors map[string]struct{}) []CheckResponse {
	keys := map[string]string{
		"next":         "unknown-task",
		"target":       "unknown-task",
		"roi":          "unknown-task",
		"entry":        "unknown-task",
		"color_filter": "unknown-task",
		"all_of":       "unknown-task",
		"any_of":       "unknown-task",
	}
	diags := make([]CheckResponse, 0)
	for key, code := range keys {
		v, ok := taskObj[key]
		if !ok {
			continue
		}
		refs := extractTaskRefsByField(key, v)
		for _, ref := range refs {
			if ref == "" {
				continue
			}
			if anchorName, isAnchorRef := parseAnchorRef(ref); isAnchorRef {
				if _, ok := anchors[anchorName]; ok {
					continue
				}
				line := findTaskKeyValueLine(pf, taskName, key, ref)
				diags = append(diags, newDiag("unknown-anchor", fmt.Sprintf("Unknown anchor %s", anchorName), pf.path, line, taskName))
				continue
			}
			if _, ok := allTasks[ref]; ok {
				continue
			}
			line := findTaskKeyValueLine(pf, taskName, key, ref)
			diags = append(diags, newDiag(code, fmt.Sprintf("Unknown task %s", ref), pf.path, line, taskName))
		}
	}
	return diags
}

func checkTemplateWarningsAndUnknownImage(pf pipelineFile, taskName string, taskObj map[string]any, images map[string]map[string]struct{}) []CheckResponse {
	diags := make([]CheckResponse, 0)
	for _, tpl := range extractTemplateRefs(taskObj) {
		line := findTaskKeyValueLine(pf, taskName, "template", tpl)
		norm := strings.TrimSpace(tpl)
		if strings.Contains(norm, `\\`) {
			diags = append(diags, newDiag("image-path-back-slash", "Image path contains backslash, shall use forward slash instead", pf.path, line, taskName))
		}
		if strings.HasPrefix(norm, "./") {
			diags = append(diags, newDiag("image-path-dot-slash", "Image path contains ./ , shall omit instead", pf.path, line, taskName))
		}
		if norm != "" && !strings.HasSuffix(strings.ToLower(norm), ".png") {
			diags = append(diags, newDiag("image-path-missing-png", "Image path shall not omit .png", pf.path, line, taskName))
		}
		if strings.ContainsAny(norm, "{}*$") {
			diags = append(diags, newDiag("dynamic-image", "Dynamic image path detected", pf.path, line, taskName))
			continue
		}
		if norm == "" || !strings.HasSuffix(strings.ToLower(norm), ".png") {
			continue
		}
		if pf.root == "" {
			continue
		}
		imgSet := images[pf.root]
		if len(imgSet) == 0 {
			continue
		}
		normPath := normalizeTemplatePath(norm)
		if _, ok := imgSet[normPath]; !ok {
			diags = append(diags, newDiag("unknown-image", fmt.Sprintf("Unknown image %s", norm), pf.path, line, taskName))
		}
	}
	return diags
}

func checkCustomActionRecoRules(pf pipelineFile, taskName string, taskObj map[string]any) []CheckResponse {
	diags := make([]CheckResponse, 0, 2)
	action, _ := taskObj["action"].(string)
	if strings.EqualFold(strings.TrimSpace(action), "Custom") {
		customAction, _ := taskObj["custom_action"].(string)
		if strings.TrimSpace(customAction) == "" {
			diags = append(diags, newDiag("error", "missing-custom-action: custom_action is empty", pf.path, findTaskKeyValueLine(pf, taskName, "action", action), taskName))
		}
	}
	recognition, _ := taskObj["recognition"].(string)
	if strings.EqualFold(strings.TrimSpace(recognition), "Custom") {
		customReco, _ := taskObj["custom_recognition"].(string)
		if strings.TrimSpace(customReco) == "" {
			diags = append(diags, newDiag("error", "missing-custom-recognition: custom_recognition is empty", pf.path, findTaskKeyValueLine(pf, taskName, "recognition", recognition), taskName))
		}
	}
	return diags
}

func checkUnknownAttr(pf pipelineFile, taskName string, taskObj map[string]any) []CheckResponse {
	diags := make([]CheckResponse, 0)
	allowed := map[string]map[string]struct{}{
		"next":   {"JumpBack": {}, "Anchor": {}},
		"target": {"Anchor": {}},
		"roi":    {"Anchor": {}},
	}
	for key, allow := range allowed {
		v, ok := taskObj[key]
		if !ok {
			continue
		}
		obj, ok := v.(map[string]any)
		if !ok {
			continue
		}
		for attr := range obj {
			if _, ok := allow[attr]; ok {
				continue
			}
			if attr == "name" || attr == "task" || attr == "value" || attr == "type" {
				continue
			}
			line := findTaskKeyValueLine(pf, taskName, key, "")
			diags = append(diags, newDiag("unknown-attr", fmt.Sprintf("Unknown attribute %s", attr), pf.path, line, taskName))
		}
	}
	return diags
}

func checkDuplicateNext(pf pipelineFile, taskName string, taskObj map[string]any) []CheckResponse {
	v, ok := taskObj["next"]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	seen := make(map[string]struct{})
	dups := make(map[string]struct{})
	for _, item := range arr {
		s, ok := item.(string)
		if !ok || s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			dups[s] = struct{}{}
			continue
		}
		seen[s] = struct{}{}
	}
	if len(dups) == 0 {
		return nil
	}
	diags := make([]CheckResponse, 0, len(dups))
	for d := range dups {
		line := findTaskKeyValueLine(pf, taskName, "next", d)
		diags = append(diags, newDiag("duplicate-next", fmt.Sprintf("Duplicate route %s", d), pf.path, line, taskName))
	}
	return diags
}

func extractTaskRefsByField(field string, v any) []string {
	refs := make([]string, 0)
	push := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		refs = append(refs, s)
	}

	switch strings.ToLower(field) {
	case "next", "entry", "all_of", "any_of", "color_filter":
		for _, s := range extractStringValues(v) {
			push(s)
		}
	case "target", "roi":
		obj, ok := v.(map[string]any)
		if !ok {
			break
		}
		if taskRef, ok := obj["task"]; ok {
			for _, s := range extractStringValues(taskRef) {
				push(s)
			}
		}
	}
	return refs
}

func extractStringValues(v any) []string {
	out := make([]string, 0)
	var walk func(any)
	walk = func(x any) {
		switch t := x.(type) {
		case string:
			out = append(out, strings.TrimSpace(t))
		case []any:
			for _, e := range t {
				walk(e)
			}
		case map[string]any:
			for _, e := range t {
				walk(e)
			}
		}
	}
	walk(v)
	return out
}

func extractTemplateRefs(taskObj map[string]any) []string {
	tpl, ok := taskObj["template"]
	if !ok {
		return nil
	}
	return extractStringValues(tpl)
}

func parseAnchorRef(ref string) (string, bool) {
	r := strings.TrimSpace(ref)
	if !strings.HasPrefix(strings.ToLower(r), strings.ToLower("[Anchor]")) {
		return "", false
	}
	name := strings.TrimSpace(r[len("[Anchor]"):])
	if name == "" {
		return "", false
	}
	return name, true
}

func extractStringByKey(v any, key string) []string {
	out := make([]string, 0)
	var walk func(any)
	walk = func(x any) {
		switch t := x.(type) {
		case map[string]any:
			for k, e := range t {
				if k == key {
					switch vv := e.(type) {
					case string:
						out = append(out, vv)
					case []any:
						for _, se := range vv {
							if sv, ok := se.(string); ok {
								out = append(out, sv)
							}
						}
					}
				}
				walk(e)
			}
		case []any:
			for _, e := range t {
				walk(e)
			}
		}
	}
	walk(v)
	return out
}

func collectPipelineFiles(roots []string) []string {
	seen := make(map[string]struct{})
	files := make([]string, 0, 64)
	for _, root := range roots {
		trimmed := strings.TrimSpace(root)
		if trimmed == "" {
			continue
		}
		abs, err := filepath.Abs(trimmed)
		if err != nil {
			continue
		}
		_ = filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				name := strings.ToLower(d.Name())
				if strings.HasPrefix(name, ".") || name == "node_modules" || name == "__pycache__" || name == ".venv" {
					return filepath.SkipDir
				}
				return nil
			}
			if !isPipelineFile(path) {
				return nil
			}
			norm := filepath.Clean(path)
			if _, ok := seen[norm]; ok {
				return nil
			}
			seen[norm] = struct{}{}
			files = append(files, norm)
			return nil
		})
	}
	return files
}

func isPipelineFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".json" && ext != ".jsonc" {
		return false
	}
	norm := strings.ToLower(filepath.ToSlash(path))
	return strings.Contains(norm, "/pipeline/") || strings.Contains(norm, "/tasks/") || strings.HasSuffix(norm, "/default_pipeline.json")
}

func collectImageFiles(roots []string) map[string]map[string]struct{} {
	result := make(map[string]map[string]struct{})
	for _, root := range roots {
		trimmed := strings.TrimSpace(root)
		if trimmed == "" {
			continue
		}
		abs, err := filepath.Abs(trimmed)
		if err != nil {
			continue
		}
		set := make(map[string]struct{})
		_ = filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if strings.ToLower(filepath.Ext(path)) != ".png" {
				return nil
			}
			norm := filepath.ToSlash(path)
			base := strings.ToLower(filepath.Base(norm))
			if base != "" {
				set[base] = struct{}{}
			}
			for _, dir := range []string{"/image/", "/template/"} {
				idx := strings.Index(strings.ToLower(norm), dir)
				if idx == -1 {
					continue
				}
				rel := norm[idx+len(dir):]
				rel = normalizeTemplatePath(rel)
				if rel != "" {
					set[rel] = struct{}{}
					set[strings.ToLower(filepath.Base(rel))] = struct{}{}
				}
			}
			return nil
		})
		result[filepath.Clean(abs)] = set
	}
	return result
}

func resolveOwnerRoot(filePath string, roots []string) string {
	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return ""
	}
	best := ""
	for _, r := range roots {
		absRoot, err := filepath.Abs(strings.TrimSpace(r))
		if err != nil {
			continue
		}
		absRoot = filepath.Clean(absRoot)
		if !strings.HasPrefix(strings.ToLower(absFile), strings.ToLower(absRoot)) {
			continue
		}
		if len(absRoot) > len(best) {
			best = absRoot
		}
	}
	return best
}

var hujsonLineColRe = regexp.MustCompile(`^hujson: line (\d+), column (\d+):`)

func parseHujsonLineCol(err error) (int, int) {
	matches := hujsonLineColRe.FindStringSubmatch(err.Error())
	if len(matches) != 3 {
		return 1, 1
	}
	line, lineErr := strconv.Atoi(matches[1])
	col, colErr := strconv.Atoi(matches[2])
	if lineErr != nil || colErr != nil || line < 1 || col < 1 {
		return 1, 1
	}
	return line, col
}

func parseStdJSONLineCol(err error) (int, int) {
	lineCol := regexp.MustCompile(`line (\d+), column (\d+)`).FindStringSubmatch(err.Error())
	if len(lineCol) != 3 {
		return 1, 1
	}
	line, e1 := strconv.Atoi(lineCol[1])
	col, e2 := strconv.Atoi(lineCol[2])
	if e1 != nil || e2 != nil || line < 1 || col < 1 {
		return 1, 1
	}
	return line, col
}

func buildLineStarts(content string) []int {
	starts := make([]int, 1, 64)
	starts[0] = 0
	for i := range content {
		if content[i] == '\n' {
			starts = append(starts, i+1)
		}
	}
	return starts
}

func offsetToLineCol(lineStarts []int, offset int) (int, int) {
	if len(lineStarts) == 0 {
		return 1, 1
	}
	if offset < 0 {
		offset = 0
	}
	idx := sort.Search(len(lineStarts), func(i int) bool {
		return lineStarts[i] > offset
	}) - 1
	if idx < 0 {
		idx = 0
	}
	line := idx + 1
	col := offset - lineStarts[idx] + 1
	if col < 1 {
		col = 1
	}
	return line, col
}

func findTaskLine(raw string, lineStarts []int, taskName string) string {
	start, _, ok := findTaskSpan(raw, taskName)
	if !ok {
		return "1:1"
	}
	line, col := offsetToLineCol(lineStarts, start)
	return fmt.Sprintf("%d:%d", line, col)
}

func findTaskKeyValueLine(pf pipelineFile, taskName string, key string, value string) string {
	start, end, ok := findTaskSpan(pf.raw, taskName)
	if !ok || start < 0 || end <= start || end > len(pf.raw) {
		return findKeyValueLine(pf.raw, pf.lineStarts, key, value)
	}

	segment := pf.raw[start:end]
	if value != "" {
		reKV := regexp.MustCompile(fmt.Sprintf(`(?m)"%s"\s*:\s*"%s"`, regexp.QuoteMeta(key), regexp.QuoteMeta(value)))
		if loc := reKV.FindStringIndex(segment); loc != nil {
			line, col := offsetToLineCol(pf.lineStarts, start+loc[0])
			return fmt.Sprintf("%d:%d", line, col)
		}
		reV := regexp.MustCompile(fmt.Sprintf(`(?m)"%s"`, regexp.QuoteMeta(value)))
		if loc := reV.FindStringIndex(segment); loc != nil {
			line, col := offsetToLineCol(pf.lineStarts, start+loc[0])
			return fmt.Sprintf("%d:%d", line, col)
		}
	}

	reK := regexp.MustCompile(fmt.Sprintf(`(?m)"%s"\s*:`, regexp.QuoteMeta(key)))
	if loc := reK.FindStringIndex(segment); loc != nil {
		line, col := offsetToLineCol(pf.lineStarts, start+loc[0])
		return fmt.Sprintf("%d:%d", line, col)
	}

	line, col := offsetToLineCol(pf.lineStarts, start)
	return fmt.Sprintf("%d:%d", line, col)
}

func findTaskSpan(raw string, taskName string) (int, int, bool) {
	re := regexp.MustCompile(fmt.Sprintf(`(?m)"%s"\s*:`, regexp.QuoteMeta(taskName)))
	loc := re.FindStringIndex(raw)
	if loc == nil {
		return -1, -1, false
	}
	objStart := -1
	for i := loc[1]; i < len(raw); i++ {
		if raw[i] == '{' {
			objStart = i
			break
		}
	}
	if objStart < 0 {
		return loc[0], loc[1], true
	}
	objEnd := matchBrace(raw, objStart)
	if objEnd < 0 {
		return loc[0], len(raw), true
	}
	return loc[0], objEnd + 1, true
}

func matchBrace(raw string, start int) int {
	if start < 0 || start >= len(raw) || raw[start] != '{' {
		return -1
	}
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(raw); i++ {
		ch := raw[i]
		if inString {
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
			continue
		}
		if ch == '{' {
			depth++
			continue
		}
		if ch == '}' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func findKeyValueLine(raw string, lineStarts []int, key string, value string) string {
	if value != "" {
		reKV := regexp.MustCompile(fmt.Sprintf(`(?m)"%s"\s*:\s*"%s"`, regexp.QuoteMeta(key), regexp.QuoteMeta(value)))
		loc := reKV.FindStringIndex(raw)
		if loc != nil {
			line, col := offsetToLineCol(lineStarts, loc[0])
			return fmt.Sprintf("%d:%d", line, col)
		}
	}
	reK := regexp.MustCompile(fmt.Sprintf(`(?m)"%s"\s*:`, regexp.QuoteMeta(key)))
	loc := reK.FindStringIndex(raw)
	if loc == nil {
		return "1:1"
	}
	line, col := offsetToLineCol(lineStarts, loc[0])
	return fmt.Sprintf("%d:%d", line, col)
}

func normalizeTemplatePath(path string) string {
	norm := strings.TrimSpace(path)
	norm = strings.ReplaceAll(norm, "\\", "/")
	norm = strings.TrimPrefix(norm, "./")
	norm = strings.TrimPrefix(norm, "/")
	norm = strings.ToLower(norm)
	for strings.Contains(norm, "//") {
		norm = strings.ReplaceAll(norm, "//", "/")
	}
	return norm
}

func splitLineCol(s string) (int, int) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return 1, 1
	}
	line, err1 := strconv.Atoi(parts[0])
	col, err2 := strconv.Atoi(parts[1])
	if err1 != nil || line < 1 {
		line = 1
	}
	if err2 != nil || col < 1 {
		col = 1
	}
	return line, col
}
