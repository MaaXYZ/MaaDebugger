package maaservice

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

const maxConsecutiveFailures = 10

type screenshotStats struct {
	capturedFrames    atomic.Uint64
	encodedFrames     atomic.Uint64
	broadcastedFrames atomic.Uint64
}

type frameSnapshot struct {
	img        image.Image
	capturedAt time.Time
}

type encodedFrame struct {
	data      []byte
	encodedAt time.Time
}

// ScreenshotService manages a periodic screenshot pipeline.
//
// The internal model is split into four stages:
//  1. capture   -> request screencap or reuse controller cache
//  2. cache     -> keep the latest raw frame in memory
//  3. encode    -> JPEG-encode the latest cached frame
//  4. broadcast -> deliver the encoded bytes via callback
//
// Public behavior remains compatible with the previous implementation.
// When a recognition event fires, the next capture cycle reads from CacheImage
// without issuing a fresh PostScreencap.
// If any stage fails maxConsecutiveFailures times in a row, the loop stops and
// onError is called.
type ScreenshotService struct {
	controllerSvc *ControllerService

	mu      sync.Mutex
	stopCh  chan struct{}
	running bool

	paused   atomic.Bool
	fps      atomic.Int32
	useCache atomic.Bool // set temporarily when reco fires so next capture reads CacheImage

	onFrame func(data []byte)
	onError func(reason string)

	cacheMu     sync.RWMutex
	latestFrame frameSnapshot

	encodedMu     sync.RWMutex
	latestEncoded encodedFrame

	stats screenshotStats
}

func NewScreenshotService(ctrlSvc *ControllerService) *ScreenshotService {
	s := &ScreenshotService{
		controllerSvc: ctrlSvc,
	}
	s.fps.Store(30)
	return s
}

func (s *ScreenshotService) SetOnFrame(fn func(data []byte)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onFrame = fn
}

func (s *ScreenshotService) SetOnError(fn func(reason string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onError = fn
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
	s.resetPipelineStateLocked()
	go s.loop(s.stopCh)
	log.Info().Int32("fps", s.fps.Load()).Msg("[Screenshot] loop started")
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
	log.Info().Msg("[Screenshot] loop stopped")
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

// NotifyRecoUpdate tells the capture stage to read from CacheImage on the next tick
// instead of issuing a fresh PostScreencap.
func (s *ScreenshotService) NotifyRecoUpdate() {
	s.useCache.Store(true)
}

func (s *ScreenshotService) loop(stop chan struct{}) {
	var buf bytes.Buffer
	jpegOpts := &jpeg.Options{Quality: 80}
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

		if err := s.captureLatestFrame(); err != nil {
			consecutiveFailures++
			log.Warn().Err(err).Int("failures", consecutiveFailures).Msg("[Screenshot] capture stage failed")
			if consecutiveFailures >= maxConsecutiveFailures {
				log.Error().Int("failures", consecutiveFailures).Msg("[Screenshot] too many consecutive failures, stopping")
				s.stopWithError(fmt.Sprintf("capture failed %d times consecutively", consecutiveFailures))
				return
			}
			continue
		}

		if err := s.encodeLatestFrame(&buf, jpegOpts); err != nil {
			consecutiveFailures++
			log.Warn().Err(err).Int("failures", consecutiveFailures).Msg("[Screenshot] encode stage failed")
			if consecutiveFailures >= maxConsecutiveFailures {
				log.Error().Int("failures", consecutiveFailures).Msg("[Screenshot] too many consecutive failures, stopping")
				s.stopWithError("JPEG encode failed repeatedly")
				return
			}
			continue
		}

		s.broadcastLatestEncoded()
		consecutiveFailures = 0
	}
}

func (s *ScreenshotService) captureLatestFrame() error {
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

	img, err := ctrl.CacheImage()
	if err != nil {
		return fmt.Errorf("CacheImage failed: %w", err)
	}

	s.cacheMu.Lock()
	s.latestFrame = frameSnapshot{
		img:        img,
		capturedAt: time.Now(),
	}
	s.cacheMu.Unlock()
	s.stats.capturedFrames.Add(1)
	return nil
}

func (s *ScreenshotService) encodeLatestFrame(buf *bytes.Buffer, jpegOpts *jpeg.Options) error {
	s.cacheMu.RLock()
	frame := s.latestFrame
	s.cacheMu.RUnlock()

	if frame.img == nil {
		return nil
	}

	buf.Reset()
	if err := jpeg.Encode(buf, frame.img, jpegOpts); err != nil {
		return err
	}

	data := append([]byte(nil), buf.Bytes()...)

	s.encodedMu.Lock()
	s.latestEncoded = encodedFrame{
		data:      data,
		encodedAt: time.Now(),
	}
	s.encodedMu.Unlock()
	s.stats.encodedFrames.Add(1)
	return nil
}

func (s *ScreenshotService) broadcastLatestEncoded() {
	s.encodedMu.RLock()
	frame := s.latestEncoded
	s.encodedMu.RUnlock()

	if len(frame.data) == 0 {
		return
	}

	s.mu.Lock()
	fn := s.onFrame
	s.mu.Unlock()
	if fn != nil {
		fn(frame.data)
		s.stats.broadcastedFrames.Add(1)
	}
}

func (s *ScreenshotService) resetPipelineStateLocked() {
	s.cacheMu.Lock()
	s.latestFrame = frameSnapshot{}
	s.cacheMu.Unlock()

	s.encodedMu.Lock()
	s.latestEncoded = encodedFrame{}
	s.encodedMu.Unlock()

	s.stats.capturedFrames.Store(0)
	s.stats.encodedFrames.Store(0)
	s.stats.broadcastedFrames.Store(0)
}

// stopWithError marks the loop as stopped and invokes the onError callback.
func (s *ScreenshotService) stopWithError(reason string) {
	s.mu.Lock()
	s.running = false
	fn := s.onError
	s.mu.Unlock()
	if fn != nil {
		fn(reason)
	}
}
