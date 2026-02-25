package maaservice

import (
	"encoding/json"
	"fmt"
	"sync"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

// TaskerService 管理 MaaFW Tasker 实例的生命周期。
// 参考 maa-js server 中的状态机：idle → running → success/failed
type TaskerService struct {
	mu     sync.Mutex
	tasker *maa.Tasker

	controllerSvc *ControllerService
	resourceSvc   *ResourceService
}

// NewTaskerService 创建一个新的 TaskerService。
func NewTaskerService(ctrlSvc *ControllerService, resSvc *ResourceService) *TaskerService {
	return &TaskerService{
		controllerSvc: ctrlSvc,
		resourceSvc:   resSvc,
	}
}

// RunTaskResult 表示任务运行结果。
type RunTaskResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// RunTask 运行指定的 task entry。
// 状态机：idle → running → success/failed
// 不在 sink 中做任何渲染/回调，避免影响运行速度。
func (s *TaskerService) RunTask(entry string, pipelineOverride json.RawMessage) RunTaskResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Info().Str("entry", entry).Msg("[MaaService] RunTask called")

	ctrl := s.controllerSvc.Controller()
	if ctrl == nil {
		log.Warn().Msg("[MaaService] RunTask: controller is not connected")
		return RunTaskResult{Error: "Controller is not connected"}
	}

	res := s.resourceSvc.Resource()
	if res == nil {
		log.Warn().Msg("[MaaService] RunTask: resource is not loaded")
		return RunTaskResult{Error: "Resource is not loaded"}
	}

	// 创建或复用 Tasker
	if s.tasker == nil {
		tasker, err := maa.NewTasker()
		if err != nil {
			log.Error().Err(err).Msg("[MaaService] create tasker failed")
			return RunTaskResult{Error: fmt.Sprintf("Failed to create tasker: %v", err)}
		}
		s.tasker = tasker
	}

	if err := s.tasker.BindController(ctrl); err != nil {
		log.Error().Err(err).Msg("[MaaService] bind controller failed")
		return RunTaskResult{Error: fmt.Sprintf("Failed to bind controller: %v", err)}
	}

	if err := s.tasker.BindResource(res); err != nil {
		log.Error().Err(err).Msg("[MaaService] bind resource failed")
		return RunTaskResult{Error: fmt.Sprintf("Failed to bind resource: %v", err)}
	}

	if !s.tasker.Initialized() {
		log.Warn().Msg("[MaaService] RunTask: tasker not initialized")
		return RunTaskResult{Error: "Failed to initialize tasker"}
	}

	// 解析 pipeline override
	var override interface{}
	if len(pipelineOverride) > 0 {
		if err := json.Unmarshal(pipelineOverride, &override); err != nil {
			log.Warn().Err(err).Msg("[MaaService] invalid pipeline override JSON")
			override = nil
		}
	}

	log.Info().Str("entry", entry).Msg("[MaaService] posting task...")
	var job *maa.TaskJob
	if override != nil {
		job = s.tasker.PostTask(entry, override)
	} else {
		job = s.tasker.PostTask(entry)
	}

	job.Wait()

	succeeded := job.Success()
	log.Info().Bool("succeeded", succeeded).Str("entry", entry).Msg("[MaaService] task completed")

	if !succeeded {
		return RunTaskResult{Error: fmt.Sprintf("Task '%s' failed", entry)}
	}

	return RunTaskResult{Success: true}
}

// StopTask 停止当前正在运行的任务。
func (s *TaskerService) StopTask() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tasker == nil {
		log.Info().Msg("[MaaService] StopTask: no active tasker")
		return
	}

	log.Info().Msg("[MaaService] stopping task...")
	job := s.tasker.PostStop()
	job.Wait()
	log.Info().Msg("[MaaService] task stopped")
}

// GetNodeList 返回当前 Resource 的节点列表。
func (s *TaskerService) GetNodeList() []string {
	res := s.resourceSvc.Resource()
	if res == nil {
		return []string{}
	}

	nodes, err := res.GetNodeList()
	if err != nil {
		log.Warn().Err(err).Msg("[MaaService] get node list failed")
		return []string{}
	}

	return nodes
}

// Running 返回 Tasker 是否正在运行任务。
func (s *TaskerService) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tasker == nil {
		return false
	}
	return s.tasker.Running()
}

// Destroy 销毁 Tasker 实例。
func (s *TaskerService) Destroy() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tasker != nil {
		s.tasker.Destroy()
		s.tasker = nil
		log.Info().Msg("[MaaService] tasker destroyed")
	}
}
