// Package configstore 提供基于本地 JSON 文件的持久化配置存储。
//
// 数据保存在 cwd/.maa/dbg.json 中，结构为 map[string]any，
// 与前端 Pinia serverPersistPlugin 的 config API 对应。
package configstore

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	dirName  = ".maa"
	fileName = "dbg.json"
)

// Store 是线程安全的、带本地文件持久化的 KV 配置存储。
type Store struct {
	mu       sync.RWMutex
	data     map[string]any
	filePath string

	// 防抖写入
	saveCh chan struct{}
	done   chan struct{}
}

// New 创建一个 Store，以 baseDir 为基目录（通常为 cwd）。
// 启动时自动从 baseDir/.maa/dbg.json 加载数据。
func New(baseDir string) *Store {
	fp := filepath.Join(baseDir, dirName, fileName)

	s := &Store{
		data:     make(map[string]any),
		filePath: fp,
		saveCh:   make(chan struct{}, 1),
		done:     make(chan struct{}),
	}

	s.loadFromDisk()

	// 启动后台防抖写入 goroutine
	go s.saveLoop()

	return s
}

// Get 返回 key 对应的值，不存在则返回 nil, false。
func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// GetAll 返回所有配置数据的浅拷贝。
func (s *Store) GetAll() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dup := make(map[string]any, len(s.data))
	maps.Copy(dup, s.data)
	return dup
}

// Set 设置单个 key 的值，并触发异步持久化。
func (s *Store) Set(key string, value any) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()

	s.scheduleSave()
}

// Merge 批量合并多个 key-value，并触发异步持久化。
func (s *Store) Merge(entries map[string]any) {
	s.mu.Lock()
	maps.Copy(s.data, entries)
	s.mu.Unlock()

	s.scheduleSave()
}

// Close 等待所有挂起的写入完成后关闭后台 goroutine。
func (s *Store) Close() {
	close(s.saveCh)
	<-s.done
}

// ---------- 内部方法 ----------

// loadFromDisk 从文件加载数据。文件不存在时静默返回。
func (s *Store) loadFromDisk() {
	raw, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info().Str("path", s.filePath).Msg("[ConfigStore] no existing config file, starting fresh")
			return
		}
		log.Error().Err(err).Str("path", s.filePath).Msg("[ConfigStore] failed to read config file")
		return
	}

	if len(raw) == 0 {
		return
	}

	var loaded map[string]any
	if err := json.Unmarshal(raw, &loaded); err != nil {
		log.Error().Err(err).Str("path", s.filePath).Msg("[ConfigStore] failed to parse config file")
		return
	}

	s.data = loaded
	log.Info().Int("keys", len(loaded)).Str("path", s.filePath).Msg("[ConfigStore] loaded config from disk")
}

// scheduleSave 通知后台 goroutine 需要写盘（非阻塞）。
func (s *Store) scheduleSave() {
	select {
	case s.saveCh <- struct{}{}:
	default:
		// 已有待处理的保存信号，无需重复发送
	}
}

// saveLoop 后台防抖写入循环，合并 500ms 内的多次写请求为一次磁盘写入。
func (s *Store) saveLoop() {
	defer close(s.done)

	for range s.saveCh {
		// 收到信号后等待一段时间（防抖），合并连续写入
		time.Sleep(500 * time.Millisecond)

		// 排空信号通道中的多余信号
		drained := true
		for drained {
			select {
			case _, ok := <-s.saveCh:
				if !ok {
					// 通道已关闭，执行最终保存
					s.writeToDisk()
					return
				}
			default:
				drained = false
			}
		}

		s.writeToDisk()
	}
}

// writeToDisk 将当前数据序列化写入文件。
func (s *Store) writeToDisk() {
	s.mu.RLock()
	raw, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()

	if err != nil {
		log.Error().Err(err).Msg("[ConfigStore] failed to marshal config data")
		return
	}

	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Error().Err(err).Str("dir", dir).Msg("[ConfigStore] failed to create config directory")
		return
	}

	// 原子写入：先写临时文件再重命名
	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, raw, 0o644); err != nil {
		log.Error().Err(err).Str("path", tmpPath).Msg("[ConfigStore] failed to write temp file")
		return
	}

	if err := os.Rename(tmpPath, s.filePath); err != nil {
		log.Error().Err(err).Str("from", tmpPath).Str("to", s.filePath).Msg("[ConfigStore] failed to rename temp file")
		return
	}

	log.Debug().Str("path", s.filePath).Msg("[ConfigStore] config saved to disk")
}
