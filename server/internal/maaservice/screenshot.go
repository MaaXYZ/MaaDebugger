package maaservice

import (
	"bytes"
	"fmt"
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
	capturedFrames    atomic.Uint64
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
// The internal model is split into four stages:
//  1. capture   -> request screencap or reuse controller cache
//  2. notify    -> publish a frame update event
//  3. encode    -> JPEG-encode the latest controller cache image
//  4. broadcast -> deliver the encoded bytes via callback
//
// Public behavior remains compatible with the previous implementation.
// When a recognition event fires, the next capture tick skips PostScreencap so
// downstream consumers read the controller's current cache image first.
// If any stage fails maxConsecutiveFailures times in a row, the loop stops and
// onError is called.
type frameHandler func(data []byte)
type errorHandler func(reason string)

type ScreenshotService struct {
	controllerSvc *ControllerService

	mu      sync.Mutex
	stopCh  chan struct{}
	running bool

	jpegJobs chan *jpegJob

	paused       atomic.Bool
	fps          atomic.Int32
	useCache     atomic.Bool
	frameSeq     atomic.Uint64
	outputDemand atomic.Uint32

	frameNotify atomic.Value
	onFrame     atomic.Value
	onError     atomic.Value

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

func (s *ScreenshotService) SetOutputDemand(outputs ScreenshotOutput) ScreenshotOutput {
	s.outputDemand.Store(uint32(outputs))
	return s.OutputDemand()
}

func (s *ScreenshotService) shouldEncodeJPEG() bool {
	return s.shouldEncodeOutput(ScreenshotOutputJPEG)
}

func (s *ScreenshotService) shouldEncodeOutput(output ScreenshotOutput) bool {
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

// Start begins the screenshot loop. Safe to call multiple times; only one loop runs.
func (s *ScreenshotService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.frameNotify.Store(make(chan struct{}, 1))
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
	s.paused.Store(true)
	log.Info().Msg("[Screenshot] paused")
}

func (s *ScreenshotService) Resume() {
	s.paused.Store(false)
	log.Info().Msg("[Screenshot] resumed")
}

func (s *ScreenshotService) Paused() bool {
	return s.paused.Load()
}

// NotifyRecoUpdate marks the next capture tick to skip PostScreencap so frame
// consumers observe the controller's current cache image first.
func (s *ScreenshotService) NotifyRecoUpdate() {
	s.useCache.Store(true)
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

		if err := s.triggerCapture(); err != nil {
			consecutiveFailures++
			log.Warn().Err(err).Int("failures", consecutiveFailures).Msg("[Screenshot] capture stage failed")
			if consecutiveFailures >= maxConsecutiveFailures {
				log.Error().Int("failures", consecutiveFailures).Msg("[Screenshot] too many consecutive failures, stopping")
				s.stopWithError(fmt.Sprintf("capture failed %d times consecutively", consecutiveFailures))
				return
			}
			continue
		}

		seq := s.frameSeq.Add(1)
		if s.shouldEncodeJPEG() {
			job, err := s.prepareLatestJPEGJob(seq)
			if err != nil {
				consecutiveFailures++
				log.Warn().Err(err).Uint64("seq", seq).Int("failures", consecutiveFailures).Msg("[Screenshot] jpeg snapshot stage failed")
				if consecutiveFailures >= maxConsecutiveFailures {
					log.Error().Int("failures", consecutiveFailures).Msg("[Screenshot] too many consecutive failures, stopping")
					s.stopWithError("JPEG snapshot failed repeatedly")
					return
				}
				continue
			}
			if job != nil {
				s.enqueueLatestJPEGJob(job)
			}
		}

		s.notifyFrameUpdated()
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

func (s *ScreenshotService) triggerCapture() error {
	ctrl := s.controllerSvc.Controller()
	if ctrl == nil {
		return nil
	}

	wantCache := s.useCache.Swap(false)
	if !wantCache {
		job := ctrl.PostScreencap()
		job.Wait()
		if !job.Success() {
			return fmt.Errorf("PostScreencap failed")
		}
	}

	s.stats.capturedFrames.Add(1)
	return nil
}

func (s *ScreenshotService) notifyFrameUpdated() {
	frameNotify, _ := s.frameNotify.Load().(chan struct{})
	if frameNotify == nil {
		return
	}

	select {
	case frameNotify <- struct{}{}:
	default:
	}
}

func (s *ScreenshotService) prepareLatestJPEGJob(seq uint64) (*jpegJob, error) {
	ctrl := s.controllerSvc.Controller()
	if ctrl == nil {
		return nil, nil
	}

	img, err := ctrl.CacheImage()
	if err != nil {
		return nil, fmt.Errorf("CacheImage failed: %w", err)
	}
	if img == nil {
		return nil, nil
	}

	return &jpegJob{seq: seq, img: img}, nil
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
	s.stats.capturedFrames.Store(0)
	s.stats.encodedFrames.Store(0)
	s.stats.broadcastedFrames.Store(0)
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
