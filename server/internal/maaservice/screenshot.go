package maaservice

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

const maxConsecutiveFailures = 10

// ScreenshotService manages a periodic screenshot loop.
// It captures frames from the controller and broadcasts them as JPEG binary via a callback.
// When a recognition event fires, it reads the cached image instead of doing a fresh screencap.
// If screencap fails maxConsecutiveFailures times in a row, the loop stops and onError is called.
type ScreenshotService struct {
	controllerSvc *ControllerService

	mu       sync.Mutex
	stopCh   chan struct{}
	running  bool
	paused   atomic.Bool
	fps      atomic.Int32
	useCache atomic.Bool // set temporarily when reco fires so next tick reads CacheImage

	onFrame func(data []byte)
	onError func(reason string)
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

// NotifyRecoUpdate tells the loop to read from CacheImage on the next tick
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

		// Compensate for actual frame processing time
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

		ctrl := s.controllerSvc.Controller()
		if ctrl == nil {
			continue
		}

		wantCache := s.useCache.Swap(false)

		if !wantCache {
			job := ctrl.PostScreencap()
			job.Wait()
			if !job.Success() {
				consecutiveFailures++
				log.Warn().Int("failures", consecutiveFailures).Msg("[Screenshot] PostScreencap failed")
				if consecutiveFailures >= maxConsecutiveFailures {
					log.Error().Int("failures", consecutiveFailures).Msg("[Screenshot] too many consecutive failures, stopping")
					s.stopWithError(fmt.Sprintf("screencap failed %d times consecutively", consecutiveFailures))
					return
				}
				continue
			}
		}

		img, err := ctrl.CacheImage()
		if err != nil {
			consecutiveFailures++
			log.Warn().Err(err).Int("failures", consecutiveFailures).Msg("[Screenshot] CacheImage failed")
			if consecutiveFailures >= maxConsecutiveFailures {
				log.Error().Int("failures", consecutiveFailures).Msg("[Screenshot] too many consecutive failures, stopping")
				s.stopWithError(fmt.Sprintf("CacheImage failed %d times consecutively", consecutiveFailures))
				return
			}
			continue
		}

		buf.Reset()
		if err := jpeg.Encode(&buf, img, jpegOpts); err != nil {
			consecutiveFailures++
			log.Warn().Err(err).Int("failures", consecutiveFailures).Msg("[Screenshot] JPEG encode failed")
			if consecutiveFailures >= maxConsecutiveFailures {
				log.Error().Int("failures", consecutiveFailures).Msg("[Screenshot] too many consecutive failures, stopping")
				s.stopWithError("JPEG encode failed repeatedly")
				return
			}
			continue
		}

		// Success — reset failure counter
		consecutiveFailures = 0

		s.mu.Lock()
		fn := s.onFrame
		s.mu.Unlock()
		if fn != nil {
			fn(buf.Bytes())
		}
	}
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
