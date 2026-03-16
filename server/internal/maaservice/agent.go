package maaservice

import (
	"fmt"
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

	s.Disconnect(identifier)

	res := s.resourceSvc.Resource()
	if res == nil {
		agentServiceLog.Info().Msg("no resource loaded, creating empty resource")
		var err error
		res, err = maa.NewResource()
		if err != nil {
			agentServiceLog.Error().Err(err).Msg("create empty resource failed")
			return AgentConnectResult{Error: fmt.Sprintf("create resource failed: %v", err)}
		}
	}

	var opts []maa.AgentClientOption
	if identifier != "" {
		opts = append(opts, maa.WithIdentifier(identifier))
	}

	client, err := maa.NewAgentClient(opts...)
	if err != nil {
		agentServiceLog.Error().Err(err).Str("identifier", identifier).Msg("create agent client failed")
		return AgentConnectResult{Error: fmt.Sprintf("create agent client failed: %v", err)}
	}

	if err := client.BindResource(res); err != nil {
		agentServiceLog.Error().Err(err).Str("identifier", identifier).Msg("bind resource failed")
		client.Destroy()
		return AgentConnectResult{Error: fmt.Sprintf("bind resource failed: %v", err)}
	}

	entry := &agentEntry{client: client, status: "connecting"}
	s.clients[identifier] = entry

	client.SetTimeout(5000 * time.Millisecond)
	if err := client.Connect(); err != nil {
		entry.status = "failed"
		agentServiceLog.Warn().Err(err).Str("identifier", identifier).Msg("connect failed")
		return AgentConnectResult{Error: fmt.Sprintf("connect failed: %v", err)}
	}

	if !client.Connected() {
		entry.status = "failed"
		agentServiceLog.Warn().Str("identifier", identifier).Msg("connected returned false")
		return AgentConnectResult{Error: "agent client reports not connected"}
	}

	entry.status = "connected"
	agentServiceLog.Info().Str("identifier", identifier).Msg("agent connected")
	return AgentConnectResult{Success: true}
}

func (s *AgentService) Disconnect(identifier string) {
	entry, ok := s.clients[identifier]
	if !ok {
		return
	}
	delete(s.clients, identifier)

	if entry.client != nil {
		agentServiceLog.Info().Str("identifier", identifier).Msg("disconnecting agent")
		entry.client.Destroy()
	}
}

func (s *AgentService) DisconnectAll() {
	for identifier, entry := range s.clients {
		if entry.client != nil {
			agentServiceLog.Info().Str("identifier", identifier).Msg("disconnecting agent during cleanup")
			entry.client.Destroy()
		}
	}
	s.clients = make(map[string]*agentEntry)
}

func (s *AgentService) List() []AgentInfo {
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
	if entry, ok := s.clients[identifier]; ok {
		return entry.client
	}
	return nil
}

func (s *AgentService) ConnectedCount() int {
	n := 0
	for _, entry := range s.clients {
		if entry.status == "connected" {
			n++
		}
	}
	return n
}
