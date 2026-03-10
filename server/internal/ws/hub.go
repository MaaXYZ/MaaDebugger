package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string `json:"type"`
	ID      string `json:"id,omitempty"`
	Payload any    `json:"payload"`
}

const sendBufSize = 256

type envelope struct {
	msgType int
	data    []byte
}

type Client struct {
	conn *websocket.Conn
	send chan envelope
	hub  *Hub
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for env := range c.send {
		if err := c.conn.WriteMessage(env.msgType, env.data); err != nil {
			return
		}
	}
}

// SendText enqueues a text message. Drops silently if the buffer is full.
func (c *Client) SendText(data []byte) {
	select {
	case c.send <- envelope{websocket.TextMessage, data}:
	default:
	}
}

// SendJSON marshals msg to JSON and enqueues it as a text message.
func (c *Client) SendJSON(msg Message) {
	buf, err := json.Marshal(msg)
	if err != nil {
		return
	}
	c.SendText(buf)
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*Client]struct{})}
}

// Register creates a Client for the connection, starts its writePump,
// and adds it to the hub. The caller should call Remove when done.
func (h *Hub) Register(conn *websocket.Conn) *Client {
	c := &Client{
		conn: conn,
		send: make(chan envelope, sendBufSize),
		hub:  h,
	}
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	go c.writePump()
	return c
}

func (h *Hub) Remove(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
	h.mu.Unlock()
}

// Close 关闭所有客户端连接并清空 hub。
func (h *Hub) Close() {
	h.mu.Lock()
	for c := range h.clients {
		delete(h.clients, c)
		close(c.send)
	}
	h.mu.Unlock()
}

func (h *Hub) BroadcastJSON(msg Message) {
	buf, err := json.Marshal(msg)
	if err != nil {
		return
	}

	env := envelope{websocket.TextMessage, buf}
	h.mu.RLock()
	for c := range h.clients {
		select {
		case c.send <- env:
		default:
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) BroadcastBinary(data []byte) {
	env := envelope{websocket.BinaryMessage, data}
	h.mu.RLock()
	for c := range h.clients {
		select {
		case c.send <- env:
		default:
		}
	}
	h.mu.RUnlock()
}
