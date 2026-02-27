package maaservice

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

// TaskerService 管理 MaaFW Tasker 实例的生命周期。
// 参考 maa-js server 中的状态机：idle → running → success/failed
type TaskerService struct {
	tasker atomic.Pointer[maa.Tasker]

	controllerSvc *ControllerService
	resourceSvc   *ResourceService

	// onEvent 广播回调，由 router 设置，用于将事件通过 WS 广播。
	// 不在 sink 中渲染/处理，只做消息转发。
	onEvent atomic.Pointer[func(msg map[string]interface{})]

	// actionScreenshots 缓存 action 开始前的截图（action_id → base64 PNG data URI）。
	actionScreenshots sync.Map
}

// NewTaskerService 创建一个新的 TaskerService。
func NewTaskerService(ctrlSvc *ControllerService, resSvc *ResourceService) *TaskerService {
	return &TaskerService{
		controllerSvc: ctrlSvc,
		resourceSvc:   resSvc,
	}
}

// SetEventCallback 设置事件广播回调。
func (s *TaskerService) SetEventCallback(fn func(msg map[string]interface{})) {
	s.onEvent.Store(&fn)
}

func (s *TaskerService) emitEvent(msg map[string]interface{}) {
	log.Info().Interface("event", msg).Msg("[MaaService] emitEvent")
	if fn := s.onEvent.Load(); fn != nil {
		(*fn)(msg)
	}
}

// eventStatusToString 将 EventStatus 转换为字符串后缀。
func eventStatusToString(status maa.EventStatus) string {
	switch status {
	case maa.EventStatusStarting:
		return "Starting"
	case maa.EventStatusSucceeded:
		return "Succeeded"
	case maa.EventStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// registerSinks 注册所有事件回调到 Tasker 上。
// 不在回调中做任何渲染，只将事件消息通过 emitEvent 发出。
func (s *TaskerService) registerSinks(tasker *maa.Tasker) {
	tasker.OnTaskerTask(func(event maa.EventStatus, detail maa.TaskerTaskDetail) {
		suffix := eventStatusToString(event)
		s.emitEvent(map[string]interface{}{
			"msg":     fmt.Sprintf("Task.%s", suffix),
			"task_id": detail.TaskID,
			"entry":   detail.Entry,
			"uuid":    detail.UUID,
		})
	})

	tasker.OnNodePipelineNodeInContext(func(_ *maa.Context, event maa.EventStatus, detail maa.NodePipelineNodeDetail) {
		suffix := eventStatusToString(event)
		s.emitEvent(map[string]interface{}{
			"msg":     fmt.Sprintf("PipelineNode.%s", suffix),
			"name":    detail.Name,
			"node_id": detail.NodeID,
		})
	})

	tasker.OnNodeRecognitionNodeInContext(func(_ *maa.Context, event maa.EventStatus, detail maa.NodeRecognitionNodeDetail) {
		suffix := eventStatusToString(event)
		s.emitEvent(map[string]interface{}{
			"msg":     fmt.Sprintf("RecognitionNode.%s", suffix),
			"name":    detail.Name,
			"node_id": detail.NodeID,
		})
	})

	tasker.OnNodeActionNodeInContext(func(_ *maa.Context, event maa.EventStatus, detail maa.NodeActionNodeDetail) {
		suffix := eventStatusToString(event)
		s.emitEvent(map[string]interface{}{
			"msg":     fmt.Sprintf("ActionNode.%s", suffix),
			"name":    detail.Name,
			"node_id": detail.NodeID,
		})
	})

	tasker.OnNodeNextListInContext(func(_ *maa.Context, event maa.EventStatus, detail maa.NodeNextListDetail) {
		suffix := eventStatusToString(event)
		list := make([]map[string]interface{}, 0, len(detail.List))
		for _, item := range detail.List {
			list = append(list, map[string]interface{}{
				"name":      item.Name,
				"jump_back": item.JumpBack,
				"anchor":    item.Anchor,
			})
		}
		s.emitEvent(map[string]interface{}{
			"msg":  fmt.Sprintf("NextList.%s", suffix),
			"name": detail.Name,
			"list": list,
		})
	})

	tasker.OnNodeRecognitionInContext(func(_ *maa.Context, event maa.EventStatus, detail maa.NodeRecognitionDetail) {
		suffix := eventStatusToString(event)
		s.emitEvent(map[string]interface{}{
			"msg":     fmt.Sprintf("Recognition.%s", suffix),
			"name":    detail.Name,
			"reco_id": detail.RecognitionID,
		})
	})

	tasker.OnNodeActionInContext(func(_ *maa.Context, event maa.EventStatus, detail maa.NodeActionDetail) {
		// 在 action 开始前截图并缓存
		if event == maa.EventStatusStarting {
			s.captureActionScreenshot(detail.ActionID)
		}

		suffix := eventStatusToString(event)
		s.emitEvent(map[string]interface{}{
			"msg":       fmt.Sprintf("Action.%s", suffix),
			"name":      detail.Name,
			"action_id": detail.ActionID,
		})
	})
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
	tasker := s.tasker.Load()
	if tasker == nil {
		newTasker, err := maa.NewTasker()
		if err != nil {
			log.Error().Err(err).Msg("[MaaService] create tasker failed")
			return RunTaskResult{Error: fmt.Sprintf("Failed to create tasker: %v", err)}
		}
		// 尝试原子设置，如果其他 goroutine 已经设置了，使用已有的
		if s.tasker.CompareAndSwap(nil, newTasker) {
			tasker = newTasker
			// 注册事件回调
			s.registerSinks(tasker)
		} else {
			// 其他 goroutine 已创建，销毁我们的并使用已有的
			newTasker.Destroy()
			tasker = s.tasker.Load()
		}
	}

	if err := tasker.BindController(ctrl); err != nil {
		log.Error().Err(err).Msg("[MaaService] bind controller failed")
		return RunTaskResult{Error: fmt.Sprintf("Failed to bind controller: %v", err)}
	}

	if err := tasker.BindResource(res); err != nil {
		log.Error().Err(err).Msg("[MaaService] bind resource failed")
		return RunTaskResult{Error: fmt.Sprintf("Failed to bind resource: %v", err)}
	}

	if !tasker.Initialized() {
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
		job = tasker.PostTask(entry, override)
	} else {
		job = tasker.PostTask(entry)
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
func (s *TaskerService) StopTask() bool {
	tasker := s.tasker.Load()
	if tasker == nil {
		log.Info().Msg("[MaaService] StopTask: no active tasker")
		return true
	}

	log.Info().Msg("[MaaService] stopping task...")
	job := tasker.PostStop()
	job.Wait()

	if job.Success() {
		log.Info().Msg("[MaaService] task stopped")
		return true
	} else {
		return false
	}
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

	sort.Strings(nodes)
	return nodes
}

// RecoResultItem 是单个识别结果（包含 box 和额外信息）。
type RecoResultItem struct {
	Box   *RectResponse `json:"box,omitempty"`
	Extra interface{}   `json:"extra,omitempty"` // score, text, count 等
}

// RecoResultsResponse 包含 all/best/filtered 三组结果。
type RecoResultsResponse struct {
	All      []*RecoResultItem `json:"all"`
	Best     []*RecoResultItem `json:"best"`
	Filtered []*RecoResultItem `json:"filtered"`
}

// RecoDetailResponse 是返回给前端的识别详情。
type RecoDetailResponse struct {
	Name           string                `json:"name"`
	Algorithm      string                `json:"algorithm"`
	Hit            bool                  `json:"hit"`
	Box            *RectResponse         `json:"box,omitempty"`
	DetailJSON     interface{}           `json:"detail_json,omitempty"`
	CombinedResult []*RecoDetailResponse `json:"combined_result,omitempty"`
	DrawImages     []string              `json:"draw_images,omitempty"`
	RawImage       string                `json:"raw_image,omitempty"`
	Results        *RecoResultsResponse  `json:"results,omitempty"`
}

// RectResponse 矩形区域。
type RectResponse struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// convertRecognitionResult 将 maa.RecognitionResult 转换为 RecoResultItem。
func convertRecognitionResult(result *maa.RecognitionResult) *RecoResultItem {
	if result == nil {
		return nil
	}

	item := &RecoResultItem{}

	switch result.Type() {
	case maa.NodeRecognitionTypeTemplateMatch:
		if v, ok := result.AsTemplateMatch(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]interface{}{
				"score": v.Score,
			}
		}
	case maa.NodeRecognitionTypeFeatureMatch:
		if v, ok := result.AsFeatureMatch(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]interface{}{
				"count": v.Count,
			}
		}
	case maa.NodeRecognitionTypeColorMatch:
		if v, ok := result.AsColorMatch(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]interface{}{
				"count": v.Count,
			}
		}
	case maa.NodeRecognitionTypeOCR:
		if v, ok := result.AsOCR(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]interface{}{
				"text":  v.Text,
				"score": v.Score,
			}
		}
	case maa.NodeRecognitionTypeNeuralNetworkClassify:
		if v, ok := result.AsNeuralNetworkClassify(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]interface{}{
				"cls_index": v.ClsIndex,
				"label":     v.Label,
				"score":     v.Score,
			}
		}
	case maa.NodeRecognitionTypeNeuralNetworkDetect:
		if v, ok := result.AsNeuralNetworkDetect(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]interface{}{
				"cls_index": v.ClsIndex,
				"label":     v.Label,
				"score":     v.Score,
			}
		}
	case maa.NodeRecognitionTypeCustom:
		if v, ok := result.AsCustom(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]interface{}{
				"detail": v.Detail,
			}
		}
	}

	return item
}

// convertResultList 将 []*maa.RecognitionResult 转换为 []*RecoResultItem。
func convertResultList(results []*maa.RecognitionResult) []*RecoResultItem {
	items := make([]*RecoResultItem, 0, len(results))
	for _, r := range results {
		if item := convertRecognitionResult(r); item != nil {
			items = append(items, item)
		}
	}
	return items
}

// convertRecoDetail 递归转换 RecognitionDetail 到响应结构。
func convertRecoDetail(detail *maa.RecognitionDetail) *RecoDetailResponse {
	if detail == nil {
		return nil
	}

	resp := &RecoDetailResponse{
		Name:      detail.Name,
		Algorithm: detail.Algorithm,
		Hit:       detail.Hit,
	}

	// Box
	if detail.Hit {
		resp.Box = &RectResponse{
			X: int(detail.Box.X()),
			Y: int(detail.Box.Y()),
			W: int(detail.Box.Width()),
			H: int(detail.Box.Height()),
		}
	}

	// DetailJSON
	if detail.DetailJson != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(detail.DetailJson), &parsed); err == nil {
			resp.DetailJSON = parsed
		} else {
			resp.DetailJSON = detail.DetailJson
		}
	}

	// CombinedResult (for And/Or algorithms)
	if len(detail.CombinedResult) > 0 {
		resp.CombinedResult = make([]*RecoDetailResponse, 0, len(detail.CombinedResult))
		for _, sub := range detail.CombinedResult {
			resp.CombinedResult = append(resp.CombinedResult, convertRecoDetail(sub))
		}
	}

	// Draw images → base64 PNG
	if len(detail.Draws) > 0 {
		resp.DrawImages = make([]string, 0, len(detail.Draws))
		for _, img := range detail.Draws {
			if img == nil {
				continue
			}
			var buf strings.Builder
			buf.WriteString("data:image/png;base64,")
			encoder := base64.NewEncoder(base64.StdEncoding, &buf)
			if err := png.Encode(encoder, img); err != nil {
				continue
			}
			encoder.Close()
			resp.DrawImages = append(resp.DrawImages, buf.String())
		}
	}

	// Raw image → base64 PNG (截图原图)
	if detail.Raw != nil {
		var buf strings.Builder
		buf.WriteString("data:image/png;base64,")
		encoder := base64.NewEncoder(base64.StdEncoding, &buf)
		if err := png.Encode(encoder, detail.Raw); err == nil {
			encoder.Close()
			resp.RawImage = buf.String()
		}
	}

	// Results (all/best/filtered 识别结果的 box 数据)
	if detail.Results != nil {
		var best []*RecoResultItem
		if detail.Results.Best != nil {
			if item := convertRecognitionResult(detail.Results.Best); item != nil {
				best = []*RecoResultItem{item}
			}
		}
		resp.Results = &RecoResultsResponse{
			All:      convertResultList(detail.Results.All),
			Best:     best,
			Filtered: convertResultList(detail.Results.Filtered),
		}
	}

	return resp
}

// NodeDetailResponse 返回给前端的节点详情（包含 reco + action）。
type NodeDetailResponse struct {
	Name         string              `json:"name"`
	Recognition  *RecoDetailResponse `json:"recognition,omitempty"`
	Action       *ActionDetailResp   `json:"action,omitempty"`
	RunCompleted bool                `json:"run_completed"`
}

// PointResponse 坐标点。
type PointResponse struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// ActionDetailResp 返回给前端的 action 详情。
type ActionDetailResp struct {
	Name       string        `json:"name"`
	Action     string        `json:"action"`
	Box        *RectResponse `json:"box,omitempty"`
	Success    bool          `json:"success"`
	DetailJSON interface{}   `json:"detail_json,omitempty"`
	Result     interface{}   `json:"result,omitempty"`
	RawImage   string        `json:"raw_image,omitempty"`
}

func convertPoint(p maa.Point) PointResponse {
	return PointResponse{X: p.X(), Y: p.Y()}
}

func convertPointSlice(pts []maa.Point) []PointResponse {
	out := make([]PointResponse, len(pts))
	for i, p := range pts {
		out[i] = convertPoint(p)
	}
	return out
}

// convertActionResult 将 ActionResult 转换为可 JSON 序列化的 map。
func convertActionResult(result *maa.ActionResult) interface{} {
	if result == nil {
		return nil
	}

	actionType := string(result.Type())

	if v, ok := result.AsClick(); ok {
		return map[string]interface{}{
			"type":     actionType,
			"point":    convertPoint(v.Point),
			"contact":  v.Contact,
			"pressure": v.Pressure,
		}
	}
	if v, ok := result.AsLongPress(); ok {
		return map[string]interface{}{
			"type":     actionType,
			"point":    convertPoint(v.Point),
			"duration": v.Duration,
			"contact":  v.Contact,
			"pressure": v.Pressure,
		}
	}
	if v, ok := result.AsSwipe(); ok {
		return map[string]interface{}{
			"type":       actionType,
			"begin":      convertPoint(v.Begin),
			"end":        convertPointSlice(v.End),
			"end_hold":   v.EndHold,
			"duration":   v.Duration,
			"only_hover": v.OnlyHover,
			"starting":   v.Starting,
			"contact":    v.Contact,
			"pressure":   v.Pressure,
		}
	}
	if v, ok := result.AsMultiSwipe(); ok {
		swipes := make([]interface{}, len(v.Swipes))
		for i, s := range v.Swipes {
			swipes[i] = map[string]interface{}{
				"begin":      convertPoint(s.Begin),
				"end":        convertPointSlice(s.End),
				"end_hold":   s.EndHold,
				"duration":   s.Duration,
				"only_hover": s.OnlyHover,
				"starting":   s.Starting,
				"contact":    s.Contact,
				"pressure":   s.Pressure,
			}
		}
		return map[string]interface{}{
			"type":   actionType,
			"swipes": swipes,
		}
	}
	if v, ok := result.AsTouch(); ok {
		return map[string]interface{}{
			"type":     actionType,
			"point":    convertPoint(v.Point),
			"contact":  v.Contact,
			"pressure": v.Pressure,
		}
	}
	if v, ok := result.AsScroll(); ok {
		return map[string]interface{}{
			"type":  actionType,
			"point": convertPoint(v.Point),
			"dx":    v.Dx,
			"dy":    v.Dy,
		}
	}
	if v, ok := result.AsClickKey(); ok {
		return map[string]interface{}{
			"type":    actionType,
			"keycode": v.Keycode,
		}
	}
	if v, ok := result.AsLongPressKey(); ok {
		return map[string]interface{}{
			"type":     actionType,
			"keycode":  v.Keycode,
			"duration": v.Duration,
		}
	}
	if v, ok := result.AsInputText(); ok {
		return map[string]interface{}{
			"type": actionType,
			"text": v.Text,
		}
	}
	if v, ok := result.AsApp(); ok {
		return map[string]interface{}{
			"type":    actionType,
			"package": v.Package,
		}
	}
	if v, ok := result.AsShell(); ok {
		return map[string]interface{}{
			"type":    actionType,
			"cmd":     v.Cmd,
			"timeout": v.ShellTimeout,
			"success": v.Success,
			"output":  v.Output,
		}
	}

	return map[string]interface{}{
		"type": actionType,
	}
}

// convertActionDetail 转换 ActionDetail 到响应结构。
func convertActionDetail(detail *maa.ActionDetail) *ActionDetailResp {
	if detail == nil {
		return nil
	}
	resp := &ActionDetailResp{
		Name:    detail.Name,
		Action:  detail.Action,
		Success: detail.Success,
	}
	resp.Box = &RectResponse{
		X: detail.Box.X(),
		Y: detail.Box.Y(),
		W: detail.Box.Width(),
		H: detail.Box.Height(),
	}
	if detail.DetailJson != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(detail.DetailJson), &parsed); err == nil {
			resp.DetailJSON = parsed
		} else {
			resp.DetailJSON = detail.DetailJson
		}
	}
	resp.Result = convertActionResult(detail.Result)
	return resp
}

// GetLatestNodeDetail 获取指定 task name 的最新节点详情。
func (s *TaskerService) GetLatestNodeDetail(name string) (*NodeDetailResponse, error) {
	tasker := s.tasker.Load()
	if tasker == nil {
		return nil, fmt.Errorf("tasker is not initialized")
	}

	detail, err := tasker.GetLatestNode(name)
	if err != nil {
		return nil, fmt.Errorf("get latest node failed: %w", err)
	}
	if detail == nil {
		return nil, fmt.Errorf("node detail not found for %s", name)
	}

	resp := &NodeDetailResponse{
		Name:         detail.Name,
		RunCompleted: detail.RunCompleted,
	}
	if detail.Recognition != nil {
		resp.Recognition = convertRecoDetail(detail.Recognition)
	}
	if detail.Action != nil {
		resp.Action = convertActionDetail(detail.Action)
	}
	return resp, nil
}

// GetRecognitionDetailByID 通过 reco_id 获取识别详情。
func (s *TaskerService) GetRecognitionDetailByID(recoID int64) (*RecoDetailResponse, error) {
	tasker := s.tasker.Load()
	if tasker == nil {
		return nil, fmt.Errorf("tasker is not initialized")
	}
	detail, err := tasker.GetRecognitionDetail(recoID)
	if err != nil {
		return nil, fmt.Errorf("get recognition detail failed: %w", err)
	}
	return convertRecoDetail(detail), nil
}

// captureActionScreenshot 在 action 开始前截图并缓存。
func (s *TaskerService) captureActionScreenshot(actionID uint64) {
	ctrl := s.controllerSvc.Controller()
	if ctrl == nil {
		return
	}
	job := ctrl.PostScreencap()
	job.Wait()
	if !job.Success() {
		log.Warn().Uint64("action_id", actionID).Msg("[MaaService] action screenshot: PostScreencap failed")
		return
	}
	img, err := ctrl.CacheImage()
	if err != nil || img == nil {
		log.Warn().Err(err).Uint64("action_id", actionID).Msg("[MaaService] action screenshot: CacheImage failed")
		return
	}
	var buf strings.Builder
	buf.WriteString("data:image/png;base64,")
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if err := png.Encode(encoder, img); err != nil {
		log.Warn().Err(err).Uint64("action_id", actionID).Msg("[MaaService] action screenshot: PNG encode failed")
		return
	}
	encoder.Close()
	s.actionScreenshots.Store(actionID, buf.String())
	log.Debug().Uint64("action_id", actionID).Msg("[MaaService] action screenshot captured")
}

// ClearActionScreenshots 清除所有缓存的 action 截图。
func (s *TaskerService) ClearActionScreenshots() {
	s.actionScreenshots.Range(func(key, _ any) bool {
		s.actionScreenshots.Delete(key)
		return true
	})
}

// GetActionDetailByID 通过 action_id 获取动作详情。
func (s *TaskerService) GetActionDetailByID(actionID int64) (*ActionDetailResp, error) {
	tasker := s.tasker.Load()
	if tasker == nil {
		return nil, fmt.Errorf("tasker is not initialized")
	}
	detail, err := tasker.GetActionDetail(actionID)
	if err != nil {
		return nil, fmt.Errorf("get action detail failed: %w", err)
	}
	resp := convertActionDetail(detail)

	// 从缓存中获取 action 开始前的截图
	if cached, ok := s.actionScreenshots.Load(actionID); ok {
		resp.RawImage = cached.(string)
	}

	return resp, nil
}

// Running 返回 Tasker 是否正在运行任务。
func (s *TaskerService) Running() bool {
	tasker := s.tasker.Load()
	if tasker == nil {
		return false
	}
	return tasker.Running()
}

// Destroy 销毁 Tasker 实例。
func (s *TaskerService) Destroy() {
	if old := s.tasker.Swap(nil); old != nil {
		old.Destroy()
		log.Info().Msg("[MaaService] tasker destroyed")
	}
}
