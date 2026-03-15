package maaservice

import (
	"encoding/json"
	"fmt"
	"sort"
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
	screenshotSvc *ScreenshotService

	// onEvent 广播回调，由 router 设置，用于将事件通过 WS 广播。
	// 不在 sink 中渲染/处理，只做消息转发。
	onEvent atomic.Pointer[func(msg map[string]any)]

	// taskImages 缓存 reco/action 详情图，供独立 image 接口按需读取。
	taskImages sync.Map
	// nodeDataByID 按 reco/action id 缓存运行时节点定义。
	nodeDataByID sync.Map
}

// NewTaskerService 创建一个新的 TaskerService。
func NewTaskerService(ctrlSvc *ControllerService, resSvc *ResourceService, screenshotSvc *ScreenshotService) *TaskerService {
	return &TaskerService{
		controllerSvc: ctrlSvc,
		resourceSvc:   resSvc,
		screenshotSvc: screenshotSvc,
	}
}

// SetEventCallback 设置事件广播回调。
func (s *TaskerService) SetEventCallback(fn func(msg map[string]any)) {
	s.onEvent.Store(&fn)
}

func (s *TaskerService) emitEvent(msg map[string]any) {
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

type runtimeNodeRecognitionV2 struct {
	Type string `json:"type"`
}

type runtimeNodeRecognitionSnapshot struct {
	Recognition any `json:"recognition"`
}

func resolveNextItemAlgorithm(ctx *maa.Context, name string) string {
	if ctx == nil || name == "" {
		return ""
	}

	nodeJSON, err := ctx.GetNodeJSON(name)
	if err != nil || nodeJSON == "" {
		return ""
	}

	var snapshot runtimeNodeRecognitionSnapshot
	if err := json.Unmarshal([]byte(nodeJSON), &snapshot); err != nil {
		return ""
	}

	switch recognition := snapshot.Recognition.(type) {
	case string:
		if recognition == "And" || recognition == "Or" {
			return recognition
		}
	case map[string]any:
		if rawType, ok := recognition["type"].(string); ok && (rawType == "And" || rawType == "Or") {
			return rawType
		}
	}

	var legacy struct {
		Recognition runtimeNodeRecognitionV2 `json:"recognition"`
	}
	if err := json.Unmarshal([]byte(nodeJSON), &legacy); err == nil {
		if legacy.Recognition.Type == "And" || legacy.Recognition.Type == "Or" {
			return legacy.Recognition.Type
		}
	}

	return ""
}

// registerSinks 注册所有事件回调到 Tasker 上。
// 不在回调中做任何渲染，只将事件消息通过 emitEvent 发出。
func (s *TaskerService) registerSinks(tasker *maa.Tasker) {
	tasker.OnTaskerTask(func(event maa.EventStatus, detail maa.TaskerTaskDetail) {
		suffix := eventStatusToString(event)
		s.emitEvent(map[string]any{
			"msg":     fmt.Sprintf("Task.%s", suffix),
			"task_id": detail.TaskID,
			"entry":   detail.Entry,
			"uuid":    detail.UUID,
		})
	})

	tasker.OnNodePipelineNodeInContext(func(_ *maa.Context, event maa.EventStatus, detail maa.NodePipelineNodeDetail) {
		suffix := eventStatusToString(event)
		s.emitEvent(map[string]any{
			"msg":     fmt.Sprintf("PipelineNode.%s", suffix),
			"name":    detail.Name,
			"node_id": detail.NodeID,
		})
	})

	tasker.OnNodeRecognitionNodeInContext(func(_ *maa.Context, event maa.EventStatus, detail maa.NodeRecognitionNodeDetail) {
		suffix := eventStatusToString(event)
		s.emitEvent(map[string]any{
			"msg":     fmt.Sprintf("RecognitionNode.%s", suffix),
			"name":    detail.Name,
			"node_id": detail.NodeID,
		})
	})

	tasker.OnNodeActionNodeInContext(func(_ *maa.Context, event maa.EventStatus, detail maa.NodeActionNodeDetail) {
		suffix := eventStatusToString(event)
		s.emitEvent(map[string]any{
			"msg":     fmt.Sprintf("ActionNode.%s", suffix),
			"name":    detail.Name,
			"node_id": detail.NodeID,
		})
	})

	tasker.OnNodeNextListInContext(func(ctx *maa.Context, event maa.EventStatus, detail maa.NodeNextListDetail) {
		suffix := eventStatusToString(event)
		list := make([]map[string]any, 0, len(detail.List))
		for _, item := range detail.List {
			entry := map[string]any{
				"name":      item.Name,
				"jump_back": item.JumpBack,
				"anchor":    item.Anchor,
				"label":     item.FormatName(),
			}
			if algorithm := resolveNextItemAlgorithm(ctx, item.Name); algorithm != "" {
				entry["algorithm"] = algorithm
			}
			list = append(list, entry)
		}
		s.emitEvent(map[string]any{
			"msg":  fmt.Sprintf("NextList.%s", suffix),
			"name": detail.Name,
			"list": list,
		})
	})
	tasker.OnNodeRecognitionInContext(func(ctx *maa.Context, event maa.EventStatus, detail maa.NodeRecognitionDetail) {
		s.cacheRuntimeNodeData(ctx, detail.Name, int64(detail.RecognitionID))
		if s.screenshotSvc != nil {
			s.screenshotSvc.NotifyCacheChanged()
		}

		suffix := eventStatusToString(event)
		s.emitEvent(map[string]any{
			"msg":     fmt.Sprintf("Recognition.%s", suffix),
			"name":    detail.Name,
			"reco_id": detail.RecognitionID,
		})
	})

	tasker.OnNodeActionInContext(func(ctx *maa.Context, event maa.EventStatus, detail maa.NodeActionDetail) {
		s.cacheRuntimeNodeData(ctx, detail.Name, int64(detail.ActionID))
		// 在 action 开始前截图并缓存（仅对有坐标的 action 类型）
		if event == maa.EventStatusStarting {
			node, err := ctx.GetNode(detail.Name)
			if err != nil {
				log.Error().Err(err).Msg("[MaaService] get node failed")
			} else if s.actionNeedsScreenshot(node) {
				s.captureActionScreenshot(detail.ActionID)
			}
		}

		suffix := eventStatusToString(event)
		s.emitEvent(map[string]any{
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
	var override any
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
	Extra any           `json:"extra,omitempty"` // score, text, count 等
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
	DetailJSON     any                   `json:"detail_json,omitempty"`
	CombinedResult []*RecoDetailResponse `json:"combined_result,omitempty"`
	DrawImages     []*ImageRef           `json:"draw_images,omitempty"`
	RawImage       *ImageRef             `json:"raw_image,omitempty"`
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
	case maa.RecognitionTypeTemplateMatch:
		if v, ok := result.AsTemplateMatch(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]any{
				"score": v.Score,
			}
		}
	case maa.RecognitionTypeFeatureMatch:
		if v, ok := result.AsFeatureMatch(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]any{
				"count": v.Count,
			}
		}
	case maa.RecognitionTypeColorMatch:
		if v, ok := result.AsColorMatch(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]any{
				"count": v.Count,
			}
		}
	case maa.RecognitionTypeOCR:
		if v, ok := result.AsOCR(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]any{
				"text":  v.Text,
				"score": v.Score,
			}
		}
	case maa.RecognitionTypeNeuralNetworkClassify:
		if v, ok := result.AsNeuralNetworkClassify(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]any{
				"cls_index": v.ClsIndex,
				"label":     v.Label,
				"score":     v.Score,
			}
		}
	case maa.RecognitionTypeNeuralNetworkDetect:
		if v, ok := result.AsNeuralNetworkDetect(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]any{
				"cls_index": v.ClsIndex,
				"label":     v.Label,
				"score":     v.Score,
			}
		}
	case maa.RecognitionTypeCustom:
		if v, ok := result.AsCustom(); ok {
			item.Box = &RectResponse{
				X: v.Box.X(), Y: v.Box.Y(),
				W: v.Box.Width(), H: v.Box.Height(),
			}
			item.Extra = map[string]any{
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
func (s *TaskerService) convertRecoDetail(detail *maa.RecognitionDetail) *RecoDetailResponse {
	if detail == nil {
		return nil
	}

	resp := &RecoDetailResponse{
		Name:      detail.Name,
		Algorithm: detail.Algorithm,
		Hit:       detail.Hit,
	}

	if detail.Hit {
		resp.Box = &RectResponse{
			X: int(detail.Box.X()),
			Y: int(detail.Box.Y()),
			W: int(detail.Box.Width()),
			H: int(detail.Box.Height()),
		}
	}

	if detail.DetailJson != "" {
		resp.DetailJSON = detail.DetailJson
	}

	if len(detail.CombinedResult) > 0 {
		resp.CombinedResult = make([]*RecoDetailResponse, 0, len(detail.CombinedResult))
		for _, sub := range detail.CombinedResult {
			resp.CombinedResult = append(resp.CombinedResult, s.convertRecoDetail(sub))
		}
	}

	if detail.Raw != nil {
		resp.RawImage = storeTaskImage(&s.taskImages, fmt.Sprintf("reco:raw-%d", detail.ID), detail.Raw)
	}

	if len(detail.Draws) > 0 {
		resp.DrawImages = make([]*ImageRef, 0, len(detail.Draws))
		for idx, drawImg := range detail.Draws {
			if drawImg == nil {
				continue
			}
			if ref := storeTaskImage(&s.taskImages, fmt.Sprintf("reco:draw-%d-%d", detail.ID, idx), drawImg); ref != nil {
				resp.DrawImages = append(resp.DrawImages, ref)
			}
		}
		if len(resp.DrawImages) == 0 {
			resp.DrawImages = nil
		}
	}

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

// NodeDataResponse 返回节点原始定义 JSON。
type NodeDataResponse struct {
	Name     string `json:"name"`
	NodeJSON string `json:"node_json"`
}

// PointResponse 坐标点。
type PointResponse struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// ActionDetailResp 返回给前端的 action 详情。
type ActionDetailResp struct {
	Name           string         `json:"name"`
	Action         string         `json:"action"`
	Box            *RectResponse  `json:"box,omitempty"`
	Success        bool           `json:"success"`
	DetailJSON     any            `json:"detail_json,omitempty"`
	Result         any            `json:"result,omitempty"`
	RawImage       *ImageRef      `json:"raw_image,omitempty"`
	ControllerType ControllerType `json:"controller_type"`
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
func convertActionResult(result *maa.ActionResult) any {
	if result == nil {
		return nil
	}

	actionType := string(result.Type())

	if v, ok := result.AsClick(); ok {
		return map[string]any{
			"type":     actionType,
			"point":    convertPoint(v.Point),
			"contact":  v.Contact,
			"pressure": v.Pressure,
		}
	}
	if v, ok := result.AsLongPress(); ok {
		return map[string]any{
			"type":     actionType,
			"point":    convertPoint(v.Point),
			"duration": v.Duration,
			"contact":  v.Contact,
			"pressure": v.Pressure,
		}
	}
	if v, ok := result.AsSwipe(); ok {
		return map[string]any{
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
		swipes := make([]any, len(v.Swipes))
		for i, s := range v.Swipes {
			swipes[i] = map[string]any{
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
		return map[string]any{
			"type":   actionType,
			"swipes": swipes,
		}
	}
	if v, ok := result.AsTouch(); ok {
		return map[string]any{
			"type":     actionType,
			"point":    convertPoint(v.Point),
			"contact":  v.Contact,
			"pressure": v.Pressure,
		}
	}
	if v, ok := result.AsScroll(); ok {
		return map[string]any{
			"type":  actionType,
			"point": convertPoint(v.Point),
			"dx":    v.Dx,
			"dy":    v.Dy,
		}
	}
	if v, ok := result.AsClickKey(); ok {
		return map[string]any{
			"type":    actionType,
			"keycode": v.Keycode,
		}
	}
	if v, ok := result.AsLongPressKey(); ok {
		return map[string]any{
			"type":     actionType,
			"keycode":  v.Keycode,
			"duration": v.Duration,
		}
	}
	if v, ok := result.AsInputText(); ok {
		return map[string]any{
			"type": actionType,
			"text": v.Text,
		}
	}
	if v, ok := result.AsApp(); ok {
		return map[string]any{
			"type":    actionType,
			"package": v.Package,
		}
	}
	if v, ok := result.AsShell(); ok {
		return map[string]any{
			"type":    actionType,
			"cmd":     v.Cmd,
			"timeout": v.ShellTimeout,
			"success": v.Success,
			"output":  v.Output,
		}
	}

	return map[string]any{
		"type": actionType,
	}
}

// convertActionDetail 转换 ActionDetail 到响应结构。
func convertActionDetail(detail *maa.ActionDetail, controllerType ControllerType) *ActionDetailResp {
	if detail == nil {
		return nil
	}
	resp := &ActionDetailResp{
		Name:           detail.Name,
		Action:         detail.Action,
		Success:        detail.Success,
		ControllerType: controllerType,
	}
	resp.Box = &RectResponse{
		X: detail.Box.X(),
		Y: detail.Box.Y(),
		W: detail.Box.Width(),
		H: detail.Box.Height(),
	}
	if detail.DetailJson != "" {
		var parsed any
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
		resp.Recognition = s.convertRecoDetail(detail.Recognition)
	}
	if detail.Action != nil {
		resp.Action = convertActionDetail(detail.Action, s.controllerSvc.ControllerType())
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
	return s.convertRecoDetail(detail), nil
}

func (s *TaskerService) cacheRuntimeNodeData(ctx *maa.Context, name string, id int64) {
	if ctx == nil || name == "" || id <= 0 {
		return
	}

	nodeJSON, err := ctx.GetNodeJSON(name)
	if err != nil {
		log.Warn().Err(err).Str("name", name).Int64("id", id).Msg("[MaaService] get runtime node data failed")
		return
	}

	s.nodeDataByID.Store(id, &NodeDataResponse{
		Name:     name,
		NodeJSON: nodeJSON,
	})
}

func (s *TaskerService) getCachedNodeData(id int64) (*NodeDataResponse, bool) {
	if id <= 0 {
		return nil, false
	}

	cached, ok := s.nodeDataByID.Load(id)
	if !ok {
		return nil, false
	}

	resp, ok := cached.(*NodeDataResponse)
	return resp, ok
}

// GetNodeData 获取运行时节点原始定义 JSON。
func (s *TaskerService) GetNodeData(name string, recoID, actionID int64) (*NodeDataResponse, error) {
	if recoID > 0 {
		if detail, ok := s.getCachedNodeData(recoID); ok {
			return detail, nil
		}
		return nil, fmt.Errorf("runtime node data not found for reco_id %d", recoID)
	}

	if actionID > 0 {
		if detail, ok := s.getCachedNodeData(actionID); ok {
			return detail, nil
		}
		return nil, fmt.Errorf("runtime node data not found for action_id %d", actionID)
	}

	res := s.resourceSvc.Resource()
	if res == nil {
		return nil, fmt.Errorf("resource is not loaded")
	}

	nodeJSON, err := res.GetNodeJSON(name)
	if err != nil {
		return nil, fmt.Errorf("get node data failed: %w", err)
	}

	return &NodeDataResponse{
		Name:     name,
		NodeJSON: nodeJSON,
	}, nil
}

// actionNeedsScreenshot 根据节点名从 Resource 获取 action 类型，
// 判断该 action 是否需要截图（有坐标的 action 需要截图）。
func (s *TaskerService) actionNeedsScreenshot(node *maa.Node) bool {
	switch node.Action.Type {
	case maa.ActionTypeDoNothing,
		maa.ActionTypeClickKey,
		maa.ActionTypeLongPressKey,
		maa.ActionTypeKeyDown,
		maa.ActionTypeKeyUp,
		maa.ActionTypeInputText,
		maa.ActionTypeStartApp,
		maa.ActionTypeStopApp,
		maa.ActionTypeStopTask,
		maa.ActionTypeCommand,
		maa.ActionTypeShell:
		return false

	default:
		return true
	}
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
	id := fmt.Sprintf("action:raw-%d", actionID)
	if ref := storeTaskImage(&s.taskImages, id, img); ref == nil {
		log.Warn().Uint64("action_id", actionID).Msg("[MaaService] action screenshot: JPEG encode failed")
		return
	}
	log.Debug().Uint64("action_id", actionID).Msg("[MaaService] action screenshot captured")
}

// ClearTaskImages 清除所有缓存的详情图片。
func (s *TaskerService) ClearTaskImages() {
	s.taskImages.Range(func(key, _ any) bool {
		s.taskImages.Delete(key)
		return true
	})
}

// ClearTaskerCache 清除 Tasker 缓存
func (s *TaskerService) ClearCache() {
	s.nodeDataByID.Range(func(key, _ any) bool {
		s.nodeDataByID.Delete(key)
		return true
	})

	tasker := s.tasker.Load()
	if tasker != nil {
		_ = tasker.ClearCache()
	}
}

// GetTaskImage 获取缓存图片。
func (s *TaskerService) GetTaskImage(id string) (*taskImageItem, bool) {
	item, ok := s.taskImages.Load(id)
	if !ok {
		return nil, false
	}
	imageItem, ok := item.(*taskImageItem)
	if !ok {
		return nil, false
	}
	return imageItem, true
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
	resp := convertActionDetail(detail, s.controllerSvc.ControllerType())
	if item, ok := s.GetTaskImage(fmt.Sprintf("action:raw-%d", actionID)); ok {
		resp.RawImage = buildTaskImageRef(fmt.Sprintf("action:raw-%d", actionID), item)
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
