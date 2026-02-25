package maaservice

import (
	"fmt"
	"sync"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

// ResourceService 管理 MaaFW Resource 实例的生命周期。
type ResourceService struct {
	mu       sync.Mutex
	resource *maa.Resource
}

// NewResourceService 创建一个新的 ResourceService。
func NewResourceService() *ResourceService {
	return &ResourceService{}
}

// LoadResult 表示资源加载结果。
type LoadResult struct {
	Success    bool   `json:"success"`
	FailedPath string `json:"failed_path,omitempty"`
}

// LoadBundles 逐个加载资源路径，遇到失败立即返回失败路径。
func (s *ResourceService) LoadBundles(paths []string) LoadResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Info().Strs("paths", paths).Int("count", len(paths)).Msg("[MaaService] LoadBundles called")

	// 销毁旧的 resource 实例
	if s.resource != nil {
		log.Info().Msg("[MaaService] destroying previous resource")
		s.resource.Destroy()
		s.resource = nil
	}

	res, err := maa.NewResource()
	if err != nil {
		log.Error().Err(err).Msg("[MaaService] create resource failed")
		return LoadResult{Success: false, FailedPath: "failed to create resource"}
	}

	for _, p := range paths {
		log.Info().Str("path", p).Msg("[MaaService] loading bundle...")
		job := res.PostBundle(p)
		job.Wait()

		if !job.Success() {
			log.Warn().Str("path", p).Str("status", fmt.Sprintf("%v", job.Status())).Msg("[MaaService] bundle load failed")
			res.Destroy()
			return LoadResult{Success: false, FailedPath: p}
		}
		log.Info().Str("path", p).Msg("[MaaService] bundle loaded successfully")
	}

	s.resource = res
	log.Info().Int("count", len(paths)).Msg("[MaaService] all bundles loaded successfully")
	return LoadResult{Success: true}
}

// Resource 返回当前的 Resource 实例（可能为 nil）。
func (s *ResourceService) Resource() *maa.Resource {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.resource
}

// Loaded 返回当前是否已加载资源。
func (s *ResourceService) Loaded() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.resource == nil {
		return false
	}
	return s.resource.Loaded()
}
