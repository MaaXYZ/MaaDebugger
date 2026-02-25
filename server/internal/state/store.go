package state

import "sync"

// Snapshot mirrors frontend status shape.
type Snapshot struct {
	Controller string `json:"controller"`
	Resource   string `json:"resource"`
	Task       string `json:"task"`
	Agent      string `json:"agent"`
}

type Store struct {
	mu       sync.RWMutex
	snapshot Snapshot
}

func NewStore() *Store {
	return &Store{
		snapshot: Snapshot{
			Controller: "disconnected",
			Resource:   "unloaded",
			Task:       "idle",
			Agent:      "disconnected",
		},
	}
}

func (s *Store) Get() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

func (s *Store) SetController(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.Controller = v
}

func (s *Store) SetResource(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.Resource = v
}

func (s *Store) SetTask(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.Task = v
}

func (s *Store) SetAgent(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot.Agent = v
}
