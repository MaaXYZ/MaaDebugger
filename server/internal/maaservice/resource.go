package maaservice

import (
	"fmt"
	"sync/atomic"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

// ResourceService 管理 MaaFW Resource 实例的生命周期。
type ResourceService struct {
	resource atomic.Pointer[maa.Resource]
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
	log.Info().Strs("paths", paths).Int("count", len(paths)).Msg("[MaaService] LoadBundles called")

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

	// 替换旧实例
	if old := s.resource.Swap(res); old != nil {
		old.Destroy()
		log.Info().Msg("[MaaService] previous resource destroyed")
	}

	log.Info().Int("count", len(paths)).Msg("[MaaService] all bundles loaded successfully")
	return LoadResult{Success: true}
}

// Resource 返回当前的 Resource 实例（可能为 nil）。
func (s *ResourceService) Resource() *maa.Resource {
	return s.resource.Load()
}

// Loaded 返回当前是否已加载资源。
func (s *ResourceService) Loaded() bool {
	res := s.resource.Load()
	if res == nil {
		return false
	}
	return res.Loaded()
}
