package maaservice

import (
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
	Msg   string `json:"msg"`
	Path  string `json:"path"`
	Line  string `json:"line"`
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
	defer s.mux.Unlock()
	s.paths = paths
}

func (s *PipelineChecker) Start() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.running = true
}

func (s *PipelineChecker) Stop() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.running = false
}

func (s *PipelineChecker) OnPathChanged(_ string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	running := s.running
	if !running {
		return
	}
	s.triggerAsync(false)
}

func (s *PipelineChecker) GetResult() []CheckResponse {
	s.mux.Lock()
	defer s.mux.Unlock()

	for s.checking {
		s.cond.Wait()
	}
	result := append([]CheckResponse(nil), s.cache...)
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
	defer s.mux.Unlock()

	if s.checking {
		s.pending = true
		return
	}
	s.checking = true
	paths := append([]string(nil), s.paths...)

	go func() {
		result := runPipelineCheck(paths)
		s.finishRun(result)

		s.mux.Lock()
		defer s.mux.Unlock()

		runAgain := s.pending
		s.pending = false
		if runAgain || force {
			s.triggerAsync(false)
		}
	}()
}

func (s *PipelineChecker) prepareRun() []string {
	s.mux.Lock()
	defer s.mux.Unlock()

	for s.checking {
		s.cond.Wait()
	}
	s.checking = true
	paths := append([]string(nil), s.paths...)
	return paths
}

func (s *PipelineChecker) finishRun(result []CheckResponse) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.cache = append([]CheckResponse(nil), result...)
	s.checking = false
	s.cond.Broadcast()
}

func runPipelineCheck(roots []string) []CheckResponse {
	files := collectPipelineFiles(roots)
	result := make([]CheckResponse, 0)
	for _, file := range files {
		diag := checkJSONCFile(file)
		if diag != nil {
			result = append(result, *diag)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Path == result[j].Path {
			return result[i].Line < result[j].Line
		}
		return result[i].Path < result[j].Path
	})
	return result
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

func checkJSONCFile(filePath string) *CheckResponse {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return &CheckResponse{
			Level: "error",
			Msg:   err.Error(),
			Path:  filePath,
			Line:  "1:1",
		}
	}

	if _, err := hujson.Parse(content); err != nil {
		line, col := parseHujsonLineCol(err)
		return &CheckResponse{
			Level: "error",
			Msg:   err.Error(),
			Path:  filePath,
			Line:  fmt.Sprintf("%d:%d", line, col),
		}
	}
	return nil
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
