package maaservice

import (
	"fmt"
	"sync"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"

	"github.com/MaaXYZ/MaaDebugger/internal/logger"
)

var agentServiceLog = logger.For(logger.ComponentAgent)

type AgentInfo struct {
	Identifier string `json:"identifier"`
	Status     string `json:"status"` // connecting | connected | failed
	Error      string `json:"error,omitempty"`
}

type AgentConnectResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type AgentService struct {
	mu          sync.Mutex
	clients     map[string]*agentEntry
	resourceSvc *ResourceService
}

type agentEntry struct {
	client *maa.AgentClient
	status string
}

func NewAgentService(resSvc *ResourceService) *AgentService {
	return &AgentService{
		clients:     make(map[string]*agentEntry),
		resourceSvc: resSvc,
	}
}

func (s *AgentService) Connect(identifier string) AgentConnectResult {
	agentServiceLog.Info().Str("identifier", identifier).Msg("connect request")
	if identifier == "" {
		return AgentConnectResult{Error: "identifier is empty"}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.disconnectLocked(identifier, false)

	res := s.resourceSvc.Resource()
	if res == nil {
		return AgentConnectResult{Error: "Resource is null."}
	}

	client, err := maa.NewAgentClient(maa.WithIdentifier(identifier))
	if err != nil {
		agentServiceLog.Error().Err(err).Str("identifier", identifier).Msg("create agent client failed")
		return AgentConnectResult{Error: fmt.Sprintf("create agent client failed: %v", err)}
	}

	if err := client.BindResource(res); err != nil {
		client.Destroy()
		agentServiceLog.Error().Err(err).Str("identifier", identifier).Msg("bind resource failed")
		return AgentConnectResult{Error: fmt.Sprintf("bind resource failed: %v", err)}
	}

	entry := &agentEntry{client: client, status: "connecting"}
	s.clients[identifier] = entry

	if err := client.SetTimeout(8000 * time.Millisecond); err != nil {
		entry.status = "failed"
		s.destroyEntry(entry)
		agentServiceLog.Warn().Err(err).Str("identifier", identifier).Msg("set timeout failed")
		return AgentConnectResult{Error: fmt.Sprintf("set timeout failed: %v", err)}
	}
	if err := client.Connect(); err != nil {
		entry.status = "failed"
		s.destroyEntry(entry)
		agentServiceLog.Warn().Err(err).Str("identifier", identifier).Msg("connect failed")
		return AgentConnectResult{Error: fmt.Sprintf("connect failed: %v", err)}
	}

	if !client.Connected() {
		entry.status = "failed"
		s.destroyEntry(entry)
		agentServiceLog.Warn().Str("identifier", identifier).Msg("connected returned false")
		return AgentConnectResult{Error: "agent client reports not connected"}
	}

	entry.status = "connected"
	agentServiceLog.Info().Str("identifier", identifier).Msg("agent connected")
	return AgentConnectResult{Success: true}
}

func (s *AgentService) Disconnect(identifier string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.disconnectLocked(identifier, false)
}

func (s *AgentService) disconnectLocked(identifier string, cleanup bool) {
	entry, ok := s.clients[identifier]
	if !ok {
		return
	}
	delete(s.clients, identifier)

	s.disconnectEntry(identifier, entry, cleanup)
}

func (s *AgentService) disconnectEntry(identifier string, entry *agentEntry, cleanup bool) {
	if entry == nil {
		return
	}

	if entry.client != nil {
		msg := "disconnecting agent"
		if cleanup {
			msg = "disconnecting agent during cleanup"
		}
		agentServiceLog.Info().Str("identifier", identifier).Msg(msg)

		// 仅在 alive + connected 时触发断连
		if entry.client.Alive() && entry.client.Connected() {
			if err := entry.client.Disconnect(); err != nil {
				warnMsg := "agent disconnect failed"
				if cleanup {
					warnMsg = "agent disconnect during cleanup failed"
				}
				agentServiceLog.Warn().Err(err).Str("identifier", identifier).Msg(warnMsg)
			}
		}
	}
	entry.status = "failed"
	s.destroyEntry(entry)
}

func (s *AgentService) destroyEntry(entry *agentEntry) {
	if entry == nil {
		return
	}

	if entry.client != nil {
		entry.client.Destroy()
		entry.client = nil
	}
}

func (s *AgentService) DisconnectAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for identifier, entry := range s.clients {
		s.disconnectEntry(identifier, entry, true)
	}
	s.clients = make(map[string]*agentEntry)
}

func (s *AgentService) List() []AgentInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]AgentInfo, 0, len(s.clients))
	for identifier, entry := range s.clients {
		info := AgentInfo{
			Identifier: identifier,
			Status:     entry.status,
		}
		if entry.status == "connected" && entry.client != nil && !entry.client.Alive() {
			info.Status = "failed"
			entry.status = "failed"
		}
		out = append(out, info)
	}
	return out
}

func (s *AgentService) GetClient(identifier string) *maa.AgentClient {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.clients[identifier]; ok {
		return entry.client
	}
	return nil
}

func (s *AgentService) ConnectedClients() map[string]*maa.AgentClient {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make(map[string]*maa.AgentClient)
	for identifier, entry := range s.clients {
		if entry == nil || entry.client == nil || entry.status != "connected" {
			continue
		}
		if !entry.client.Alive() {
			entry.status = "failed"
			continue
		}
		out[identifier] = entry.client
	}
	return out
}

func (s *AgentService) ConnectedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := 0
	for _, entry := range s.clients {
		if entry.status == "connected" {
			n++
		}
	}
	return n
}
