package maaservice

import (
	"fmt"
	"sync/atomic"

	maa "github.com/MaaXYZ/maa-framework-go/v4"

	"github.com/MaaXYZ/MaaDebugger/internal/logger"
)

var resourceServiceLog = logger.For(logger.ComponentResource)

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
	resourceServiceLog.Info().Strs("paths", paths).Int("count", len(paths)).Msg("load bundles request")

	res, err := maa.NewResource()
	if err != nil {
		resourceServiceLog.Error().Err(err).Msg("create resource failed")
		return LoadResult{Success: false, FailedPath: "failed to create resource"}
	}

	for _, p := range paths {
		resourceServiceLog.Info().Str("path", p).Msg("loading bundle")
		job := res.PostBundle(p)
		job.Wait()

		if !job.Success() {
			resourceServiceLog.Warn().Str("path", p).Str("status", fmt.Sprintf("%v", job.Status())).Msg("bundle load failed")
			res.Destroy()
			return LoadResult{Success: false, FailedPath: p}
		}
		resourceServiceLog.Info().Str("path", p).Msg("bundle loaded")
	}

	// 替换旧实例
	if old := s.resource.Swap(res); old != nil {
		old.Destroy()
		resourceServiceLog.Info().Msg("previous resource destroyed")
	}

	resourceServiceLog.Info().Int("count", len(paths)).Msg("all bundles loaded")
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

// Destroy 销毁当前 Resource 实例。
func (s *ResourceService) Destroy() {
	if res := s.resource.Swap(nil); res != nil {
		res.Destroy()
		resourceServiceLog.Info().Msg("resource destroyed")
	}
}
