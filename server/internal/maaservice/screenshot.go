package maaservice

import (
	"bytes"
	"image/png"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// ScreenshotService manages a periodic screenshot loop.
// It captures frames from the controller and broadcasts them as PNG binary via a callback.
// When a recognition event fires, it reads the cached image instead of doing a fresh screencap.
type ScreenshotService struct {
	controllerSvc *ControllerService

	mu       sync.Mutex
	stopCh   chan struct{}
	running  bool
	paused   atomic.Bool
	fps      atomic.Int32
	useCache atomic.Bool // set temporarily when reco fires so next tick reads CacheImage

	onFrame func(data []byte)
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
	for {
		fps := s.fps.Load()
		interval := time.Second / time.Duration(fps)

		select {
		case <-stop:
			return
		case <-time.After(interval):
		}

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
				continue
			}
		}

		img, err := ctrl.CacheImage()
		if err != nil {
			log.Debug().Err(err).Msg("[Screenshot] CacheImage failed")
			continue
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			log.Debug().Err(err).Msg("[Screenshot] PNG encode failed")
			continue
		}

		s.mu.Lock()
		fn := s.onFrame
		s.mu.Unlock()
		if fn != nil {
			fn(buf.Bytes())
		}
	}
}
