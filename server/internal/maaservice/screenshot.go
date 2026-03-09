package maaservice

import (
	"bytes"
	"fmt"
	"hash"
	"hash/fnv"
	"image"
	"image/jpeg"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

const maxConsecutiveFailures = 10

type ScreenshotOutput uint32

const (
	ScreenshotOutputJPEG ScreenshotOutput = 1
	ScreenshotOutputH264 ScreenshotOutput = 1 << 1
	ScreenshotOutputH265 ScreenshotOutput = 1 << 2
)

type screenshotStats struct {
	encodedFrames     atomic.Uint64
	broadcastedFrames atomic.Uint64
}

type jpegJob struct {
	seq uint64
	img image.Image
}

type jpegResult struct {
	seq  uint64
	data []byte
	err  error
}

// ScreenshotService manages a periodic screenshot pipeline.
//
// The internal model is split into five stages:
//  1. tick      -> run at the configured target FPS
//  2. screencap -> while connected and task is not running, refresh controller cache via PostScreencap
//  3. poll      -> read the controller cache image as the only source of truth
//  4. compare   -> sample feature-region fingerprints and skip unchanged frames
//  5. encode    -> JPEG-encode the changed cache image and broadcast it via callback
//
// Cache remains the only truth for downstream consumers. PostScreencap is only
// used by a bounded-lifetime refresh loop to update that cache while the
// controller is connected, not manually paused, and no task is running. Once a
// task enters running state, the refresh loop stops issuing PostScreencap
// immediately, but the service still keeps polling and consuming cache on each
// tick. After the task ends, the service attempts to resume PostScreencap
// automatically unless it was manually paused or stopped. If any stage fails
// maxConsecutiveFailures times in a row, the loop stops and onError is called.
type frameHandler func(data []byte)
type errorHandler func(reason string)

type ScreenshotService struct {
	controllerSvc *ControllerService

	mu      sync.Mutex
	stopCh  chan struct{}
	running bool

	jpegJobs chan *jpegJob

	paused       atomic.Bool
	manualPaused atomic.Bool
	taskRunning  atomic.Bool
	fps          atomic.Int32
	useCache     atomic.Bool
	frameSeq     atomic.Uint64
	outputDemand atomic.Uint32
	outputActive atomic.Bool

	frameChangedNotify atomic.Value
	onFrame            atomic.Value
	onError            atomic.Value

	lastCacheHash      uint64
	lastCacheHashValid bool

	stats screenshotStats
}

func NewScreenshotService(ctrlSvc *ControllerService) *ScreenshotService {
	s := &ScreenshotService{
		controllerSvc: ctrlSvc,
	}
	s.fps.Store(15)
	return s
}

func (s *ScreenshotService) SetOnFrame(fn func(data []byte)) {
	if fn == nil {
		s.onFrame.Store(frameHandler(nil))
		s.disableOutput(ScreenshotOutputJPEG)
		return
	}

	s.onFrame.Store(frameHandler(fn))
	s.enableOutput(ScreenshotOutputJPEG)
}

func (s *ScreenshotService) EnableOutput(output ScreenshotOutput) {
	s.enableOutput(output)
}

func (s *ScreenshotService) DisableOutput(output ScreenshotOutput) {
	s.disableOutput(output)
}

func (s *ScreenshotService) EnableJPEG() {
	s.EnableOutput(ScreenshotOutputJPEG)
}

func (s *ScreenshotService) DisableJPEG() {
	s.DisableOutput(ScreenshotOutputJPEG)
}

func (s *ScreenshotService) OutputDemand() ScreenshotOutput {
	return ScreenshotOutput(s.outputDemand.Load())
}

func (s *ScreenshotService) OutputEnabled(output ScreenshotOutput) bool {
	return s.shouldEncodeOutput(output)
}

func (s *ScreenshotService) OutputActive() bool {
	return s.outputActive.Load()
}

func (s *ScreenshotService) SetOutputDemand(outputs ScreenshotOutput) ScreenshotOutput {
	s.outputDemand.Store(uint32(outputs))
	return s.OutputDemand()
}

func (s *ScreenshotService) shouldEncodeJPEG() bool {
	return s.shouldEncodeOutput(ScreenshotOutputJPEG)
}

func (s *ScreenshotService) shouldEncodeOutput(output ScreenshotOutput) bool {
	if !s.outputActive.Load() {
		return false
	}
	mask := uint32(output)
	return s.outputDemand.Load()&mask != 0
}

func (s *ScreenshotService) enableOutput(output ScreenshotOutput) {
	mask := uint32(output)
	for {
		current := s.outputDemand.Load()
		updated := current | mask
		if current == updated || s.outputDemand.CompareAndSwap(current, updated) {
			return
		}
	}
}

func (s *ScreenshotService) disableOutput(output ScreenshotOutput) {
	mask := uint32(output)
	for {
		current := s.outputDemand.Load()
		updated := current &^ mask
		if current == updated || s.outputDemand.CompareAndSwap(current, updated) {
			return
		}
	}
}

func (s *ScreenshotService) SetOnError(fn func(reason string)) {
	s.onError.Store(errorHandler(fn))
}

func (s *ScreenshotService) EnableOutputDelivery() {
	s.outputActive.Store(true)
	log.Info().Msg("[Screenshot] output delivery enabled")
}

func (s *ScreenshotService) DisableOutputDelivery() {
	s.outputActive.Store(false)
	log.Info().Msg("[Screenshot] output delivery disabled")
}

// Start begins the screenshot loop. Safe to call multiple times; only one loop runs.
func (s *ScreenshotService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.frameChangedNotify.Store(make(chan struct{}, 1))
	s.resetPipelineStateLocked()
	go s.captureLoop(s.stopCh)
	go s.jpegLoop(s.stopCh)
	log.Info().Int32("fps", s.fps.Load()).Msg("[Screenshot] loops started")
}

// Stop halts the screenshot loop.
func (s *ScreenshotService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	close(s.stopCh)
	s.running = false
	log.Info().Msg("[Screenshot] loops stopped")
}

func (s *ScreenshotService) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *ScreenshotService) SetFPS(fps int32) {
	if fps < 1 {
		fps = 1
	}
	if fps > 60 {
		fps = 60
	}
	s.fps.Store(fps)
}

func (s *ScreenshotService) GetFPS() int32 {
	return s.fps.Load()
}

func (s *ScreenshotService) Pause() {
	s.manualPaused.Store(true)
	s.paused.Store(true)
	log.Info().Msg("[Screenshot] paused")
}

func (s *ScreenshotService) Resume() {
	s.manualPaused.Store(false)
	s.resumeIfAllowed()
}

func (s *ScreenshotService) OnConnected() {
	s.Start()
	s.EnableOutputDelivery()
	s.taskRunning.Store(false)
	s.resumeIfAllowed()
}

func (s *ScreenshotService) OnTaskStarted() {
	s.taskRunning.Store(true)
	log.Info().Msg("[Screenshot] task running, PostScreencap paused")
}

func (s *ScreenshotService) OnTaskEnded() {
	s.taskRunning.Store(false)
	log.Info().Msg("[Screenshot] task ended, attempting to resume PostScreencap")
	if s.manualPaused.Load() {
		return
	}
	if !s.outputActive.Load() {
		return
	}
	s.paused.Store(false)
}

func (s *ScreenshotService) Paused() bool {
	return s.paused.Load()
}

func (s *ScreenshotService) resumeIfAllowed() {
	if s.manualPaused.Load() {
		s.paused.Store(true)
		log.Info().Msg("[Screenshot] resume skipped: manually paused")
		return
	}
	if !s.outputActive.Load() {
		s.paused.Store(true)
		log.Info().Msg("[Screenshot] resume skipped: output delivery disabled")
		return
	}

	s.paused.Store(false)
	if s.taskRunning.Load() {
		log.Info().Msg("[Screenshot] resumed cache polling with PostScreencap disabled by task state")
		return
	}
	log.Info().Msg("[Screenshot] resumed")
}

func (s *ScreenshotService) captureLoop(stop <-chan struct{}) {
	lastTick := time.Now()
	consecutiveFailures := 0

	for {
		fps := s.fps.Load()
		interval := time.Second / time.Duration(fps)

		elapsed := time.Since(lastTick)
		sleep := interval - elapsed
		if sleep < time.Millisecond {
			sleep = time.Millisecond
		}

		select {
		case <-stop:
			return
		case <-time.After(sleep):
		}

		lastTick = time.Now()

		if s.paused.Load() {
			continue
		}

		seq := s.frameSeq.Add(1)
		if err := s.refreshCacheSnapshot(seq); err != nil {
			consecutiveFailures++
			log.Warn().Err(err).Uint64("seq", seq).Int("failures", consecutiveFailures).Msg("[Screenshot] screencap stage failed")
			if consecutiveFailures >= maxConsecutiveFailures {
				log.Error().Int("failures", consecutiveFailures).Msg("[Screenshot] too many consecutive failures, stopping")
				s.stopWithError("PostScreencap failed repeatedly")
				return
			}
			continue
		}

		job, changed, err := s.prepareChangedJPEGJob(seq)
		if err != nil {
			consecutiveFailures++
			log.Warn().Err(err).Uint64("seq", seq).Int("failures", consecutiveFailures).Msg("[Screenshot] cache snapshot stage failed")
			if consecutiveFailures >= maxConsecutiveFailures {
				log.Error().Int("failures", consecutiveFailures).Msg("[Screenshot] too many consecutive failures, stopping")
				s.stopWithError("Cache snapshot failed repeatedly")
				return
			}
			continue
		}
		if !changed {
			consecutiveFailures = 0
			continue
		}

		s.notifyFrameChanged()
		if job != nil && s.shouldEncodeJPEG() {
			s.enqueueLatestJPEGJob(job)
		}
		consecutiveFailures = 0
	}
}

func (s *ScreenshotService) jpegLoop(stop <-chan struct{}) {
	workerCount := s.jpegWorkerCount()
	results := make(chan jpegResult, workerCount)
	var wg sync.WaitGroup

	for range workerCount {
		wg.Go(func() {
			s.jpegWorker(stop, results)
		})
	}

	defer wg.Wait()
	consecutiveFailures := 0
	lastEmittedSeq := uint64(0)

	for {
		select {
		case <-stop:
			return
		case result := <-results:
			if result.err != nil {
				consecutiveFailures++
				log.Warn().Err(result.err).Uint64("seq", result.seq).Int("failures", consecutiveFailures).Msg("[Screenshot] encode stage failed")
				if consecutiveFailures >= maxConsecutiveFailures {
					log.Error().Int("failures", consecutiveFailures).Msg("[Screenshot] too many consecutive failures, stopping")
					s.stopWithError("JPEG encode failed repeatedly")
					return
				}
				continue
			}

			if result.seq <= lastEmittedSeq {
				continue
			}
			if len(result.data) == 0 {
				continue
			}

			lastEmittedSeq = result.seq
			s.emitJPEGFrame(result.data)
			consecutiveFailures = 0
		}
	}
}

func (s *ScreenshotService) jpegWorker(stop <-chan struct{}, results chan<- jpegResult) {
	var buf bytes.Buffer
	jpegOpts := &jpeg.Options{Quality: 80}

	for {
		select {
		case <-stop:
			return
		case job := <-s.jpegJobs:
			if job == nil {
				continue
			}
			data, err := s.encodeJPEGImage(job.img, &buf, jpegOpts)
			select {
			case <-stop:
				return
			case results <- jpegResult{seq: job.seq, data: data, err: err}:
			}
		}
	}
}

func (s *ScreenshotService) notifyFrameChanged() {
	frameChangedNotify, _ := s.frameChangedNotify.Load().(chan struct{})
	if frameChangedNotify == nil {
		return
	}

	select {
	case frameChangedNotify <- struct{}{}:
	default:
	}
}

func (s *ScreenshotService) refreshCacheSnapshot(seq uint64) error {
	if s.taskRunning.Load() {
		return nil
	}

	ctrl := s.controllerSvc.Controller()
	if ctrl == nil {
		return nil
	}

	job := ctrl.PostScreencap()
	if job == nil {
		return fmt.Errorf("PostScreencap returned nil job")
	}
	job.Wait()
	if !job.Success() {
		return fmt.Errorf("PostScreencap failed")
	}
	return nil
}

func (s *ScreenshotService) prepareChangedJPEGJob(seq uint64) (*jpegJob, bool, error) {
	ctrl := s.controllerSvc.Controller()
	if ctrl == nil {
		return nil, false, nil
	}

	img, err := ctrl.CacheImage()
	if err != nil {
		return nil, false, fmt.Errorf("CacheImage failed: %w", err)
	}
	if img == nil {
		return nil, false, nil
	}

	hash := fingerprintImage(img)
	if s.lastCacheHashValid && s.lastCacheHash == hash {
		return nil, false, nil
	}

	s.lastCacheHash = hash
	s.lastCacheHashValid = true
	return &jpegJob{seq: seq, img: img}, true, nil
}

func (s *ScreenshotService) enqueueLatestJPEGJob(job *jpegJob) {
	if job == nil || s.jpegJobs == nil {
		return
	}

	for {
		select {
		case s.jpegJobs <- job:
			return
		default:
		}

		select {
		case <-s.jpegJobs:
		default:
		}
	}
}

func (s *ScreenshotService) jpegWorkerCount() int {
	count := runtime.GOMAXPROCS(0)
	if count < 1 {
		return 1
	}
	if count > 4 {
		return 4
	}
	return count
}

func (s *ScreenshotService) encodeJPEGImage(img image.Image, buf *bytes.Buffer, jpegOpts *jpeg.Options) ([]byte, error) {
	if img == nil {
		return nil, nil
	}

	buf.Reset()
	if err := jpeg.Encode(buf, img, jpegOpts); err != nil {
		return nil, err
	}

	data := append([]byte(nil), buf.Bytes()...)
	s.stats.encodedFrames.Add(1)
	return data, nil
}
func (s *ScreenshotService) emitJPEGFrame(data []byte) {
	if len(data) == 0 {
		return
	}

	fn, _ := s.onFrame.Load().(frameHandler)
	if fn != nil {
		fn(data)
		s.stats.broadcastedFrames.Add(1)
	}
}

func (s *ScreenshotService) resetPipelineStateLocked() {
	s.frameSeq.Store(0)
	s.jpegJobs = make(chan *jpegJob, s.jpegWorkerCount())
	s.lastCacheHash = 0
	s.lastCacheHashValid = false
	s.outputActive.Store(false)
	s.manualPaused.Store(false)
	s.taskRunning.Store(false)
	s.paused.Store(true)
	s.stats.encodedFrames.Store(0)
	s.stats.broadcastedFrames.Store(0)
}

func fingerprintImage(img image.Image) uint64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return 0
	}

	h := fnv.New64a()
	var pixel [8]byte
	regions := [9][2]float64{
		{0.15, 0.15}, {0.50, 0.15}, {0.85, 0.15},
		{0.15, 0.50}, {0.50, 0.50}, {0.85, 0.50},
		{0.15, 0.85}, {0.50, 0.85}, {0.85, 0.85},
	}

	for _, region := range regions {
		cx := bounds.Min.X + int(float64(width-1)*region[0])
		cy := bounds.Min.Y + int(float64(height-1)*region[1])
		sampleFeatureRegion(h, img, cx, cy, width, height, &pixel)
	}

	return h.Sum64()
}

func sampleFeatureRegion(h hash.Hash64, img image.Image, cx, cy, width, height int, pixel *[8]byte) {
	stepX := max(1, width/32)
	stepY := max(1, height/32)
	offsets := [5][2]int{
		{0, 0},
		{-stepX, 0},
		{stepX, 0},
		{0, -stepY},
		{0, stepY},
	}

	for _, offset := range offsets {
		x := cx + offset[0]
		y := cy + offset[1]
		r, g, b, a := img.At(x, y).RGBA()
		pixel[0] = byte(r >> 8)
		pixel[1] = byte(g >> 8)
		pixel[2] = byte(b >> 8)
		pixel[3] = byte(a >> 8)
		pixel[4] = byte((r >> 8) ^ (g >> 8))
		pixel[5] = byte((g >> 8) ^ (b >> 8))
		pixel[6] = byte((a >> 8) ^ (r >> 8))
		pixel[7] = byte((x + y) & 0xff)
		_, _ = h.Write(pixel[:])
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// stopWithError marks the loop as stopped and invokes the onError callback.
func (s *ScreenshotService) stopWithError(reason string) {
	s.mu.Lock()
	if s.running {
		close(s.stopCh)
		s.running = false
	}
	s.mu.Unlock()
	fn, _ := s.onError.Load().(errorHandler)
	if fn != nil {
		fn(reason)
	}
}
