package maaservice

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/MaaXYZ/MaaDebugger/internal/logger"
)

var ignorePaths = []string{
	// git
	".git",
	// IDE
	".vscode",
	".idea",
	// JS/TS
	"node_modules",
	// Python
	"__pycache__",
}

var watchLog = logger.For(logger.ComponentResource)

// Watcher 管理文件系统的变动监听，互不干扰
type Watcher struct {
	watcher          *fsnotify.Watcher
	paths            []string
	ctx              context.Context
	cancel           context.CancelFunc
	mu               sync.Mutex
	debounceDuration time.Duration

	onChange func(path string)
	onError  func(error)

	enabled bool
}

// NewWatcher 创建一个新的 Watcher 实例
func NewWatcher(onChange func(path string), onError func(error)) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	watcher := &Watcher{
		watcher:          w,
		ctx:              ctx,
		cancel:           cancel,
		onChange:         onChange,
		onError:          onError,
		enabled:          false,
		debounceDuration: 500 * time.Millisecond,
	}

	go watcher.watchLoop()

	return watcher, nil
}

// SetEnabled 设置是否开启监听，根据前端设置控制
func (w *Watcher) SetEnabled(enabled bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.enabled = enabled
}

// SetInterval 设置防抖时间（毫秒）
func (w *Watcher) SetInterval(intervalMs int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if intervalMs > 0 {
		w.debounceDuration = time.Duration(intervalMs) * time.Millisecond
	}
}

// SetPaths 设置需要监听的根目录并自动添加子目录
func (w *Watcher) SetPaths(paths []string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 清除旧的监听
	w.removeAll()
	w.paths = paths

	// 添加新的监听
	for _, p := range paths {
		err := w.addRecursive(p)
		if err != nil {
			watchLog.Error().Err(err).Str("path", p).Msg("failed to add watch path")
		}
	}

	return nil
}

func (w *Watcher) removeAll() {
	// 获取所有正在监听的路径并移除
	for _, p := range w.watcher.WatchList() {
		w.watcher.Remove(p)
	}
}

// addRecursive 递归添加目录到监听列表
func (w *Watcher) addRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") { // 跳过 . 开头的目录
				return filepath.SkipDir
			}
			for _, ignore := range ignorePaths {
				if strings.Contains(path, ignore) {
					return filepath.SkipDir
				}
			}
			err = w.watcher.Add(path)
			if err != nil {
				watchLog.Warn().Err(err).Str("path", path).Msg("failed to watch directory")
			}
		}
		return nil
	})
}

// Destroy 停止监听并释放资源
func (w *Watcher) Destroy() {
	w.cancel()
	w.watcher.Close()
}

func (w *Watcher) watchLoop() {
	// 防抖 timer
	var timer *time.Timer

	for {
		select {
		case <-w.ctx.Done():
			return
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			watchLog.Error().Err(err).Msg("watcher error")
			if w.onError != nil {
				w.onError(err)
			}
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			w.mu.Lock()
			enabled := w.enabled
			duration := w.debounceDuration
			w.mu.Unlock()

			if !enabled {
				continue
			}

			// 如果是新建目录，需要将其添加到监听中
			if event.Has(fsnotify.Create) {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					w.mu.Lock()
					w.addRecursive(event.Name)
					w.mu.Unlock()
				}
			}

			// 防抖处理触发回调
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(duration, func() {
				// 检查根目录是否还存在
				w.mu.Lock()
				paths := w.paths
				w.mu.Unlock()

				for _, p := range paths {
					if _, err := os.Stat(p); os.IsNotExist(err) {
						if w.onError != nil {
							w.onError(err) // 根目录不再存在，触发错误
						}
						return
					}
				}

				if w.onChange != nil {
					w.onChange(event.Name)
				}
			})
		}
	}
}
