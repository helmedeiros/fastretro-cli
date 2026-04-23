// Package server provides an embedded WebSocket relay server for fastRetro
// sessions. It implements the same room-based protocol as the web app's Vite
// plugin, allowing CLI-to-CLI sessions without the web app running.
package server

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// safeConn wraps a websocket.Conn with a write mutex for concurrent safety.
type safeConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (sc *safeConn) writeMessage(data []byte) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.WriteMessage(websocket.TextMessage, data)
}

// Room holds the state for a single session.
type Room struct {
	mu       sync.RWMutex
	state    []byte                     // raw JSON of latest RetroState
	teamInfo []byte                     // raw JSON of team-info message
	clients  map[*safeConn]bool         // active connections
	takenIDs map[string]bool            // claimed participant IDs
	votes    map[string]map[string]bool // stage → set of participantIds
}

// Server manages rooms and handles HTTP/WebSocket connections.
type Server struct {
	mu       sync.RWMutex
	rooms    map[string]*Room
	upgrader websocket.Upgrader
}

// New creates a new Server instance.
func New() *Server {
	return &Server{
		rooms: make(map[string]*Room),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

const codeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateRoomCode() string {
	parts := make([]string, 3)
	for p := range parts {
		seg := make([]byte, 3)
		for i := range seg {
			seg[i] = codeChars[rand.Intn(len(codeChars))]
		}
		parts[p] = string(seg)
	}
	return strings.Join(parts, "-")
}

func (s *Server) getOrCreateRoom(code string) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()
	room, ok := s.rooms[code]
	if !ok {
		room = &Room{
			clients:  make(map[*safeConn]bool),
			takenIDs: make(map[string]bool),
			votes:    make(map[string]map[string]bool),
		}
		s.rooms[code] = room
	}
	return room
}

func (s *Server) removeRoom(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rooms, code)
}

// Handler returns an http.Handler with the API and WebSocket endpoints.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/__api/rooms", s.handleCreateRoom)
	mux.HandleFunc("/__ws/room/", s.handleWebSocket)
	return mux
}

func (s *Server) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	code := generateRoomCode()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract room code from URL: /__ws/room/{code}
	path := strings.TrimPrefix(r.URL.Path, "/__ws/room/")
	code := strings.TrimRight(path, "/")
	if code == "" {
		http.Error(w, "missing room code", http.StatusBadRequest)
		return
	}

	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	sc := &safeConn{conn: wsConn}

	room := s.getOrCreateRoom(code)

	// Add client to room
	room.mu.Lock()
	room.clients[sc] = true
	room.mu.Unlock()

	// Send initial state to new client
	s.sendInitialState(room, sc)

	// Broadcast updated peer count
	s.broadcastPeerCount(room)

	// Read loop
	defer func() {
		room.mu.Lock()
		delete(room.clients, sc)
		empty := len(room.clients) == 0
		room.mu.Unlock()

		wsConn.Close()
		s.broadcastPeerCount(room)

		if empty {
			s.removeRoom(code)
		}
	}()

	for {
		_, raw, err := wsConn.ReadMessage()
		if err != nil {
			break
		}
		s.routeMessage(room, sc, raw)
	}
}

func (s *Server) sendInitialState(room *Room, sc *safeConn) {
	room.mu.RLock()
	defer room.mu.RUnlock()

	if room.state != nil {
		_ = sc.writeMessage(room.state)
	}

	ids := make([]string, 0, len(room.takenIDs))
	for id := range room.takenIDs {
		ids = append(ids, id)
	}
	if msg, err := json.Marshal(map[string]interface{}{"type": "taken-ids", "ids": ids}); err == nil {
		_ = sc.writeMessage(msg)
	}

	if room.teamInfo != nil {
		_ = sc.writeMessage(room.teamInfo)
	}
}

func (s *Server) broadcastPeerCount(room *Room) {
	room.mu.RLock()
	count := len(room.clients)
	room.mu.RUnlock()

	msg, _ := json.Marshal(map[string]interface{}{"type": "peer-count", "count": count})
	s.broadcast(room, msg, nil)
}

func (s *Server) broadcast(room *Room, data []byte, exclude *safeConn) {
	room.mu.RLock()
	defer room.mu.RUnlock()

	for client := range room.clients {
		if client != exclude {
			_ = client.writeMessage(data)
		}
	}
}

type incomingMsg struct {
	Type          string `json:"type"`
	ParticipantID string `json:"participantId,omitempty"`
	Stage         string `json:"stage,omitempty"`
}

func (s *Server) routeMessage(room *Room, sender *safeConn, raw []byte) {
	var msg incomingMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return
	}

	switch msg.Type {
	case "state":
		room.mu.Lock()
		room.state = raw
		room.mu.Unlock()
		s.broadcast(room, raw, sender)

	case "team-info":
		room.mu.Lock()
		room.teamInfo = raw
		room.mu.Unlock()
		s.broadcast(room, raw, nil)

	case "claim-identity":
		room.mu.Lock()
		if msg.ParticipantID != "" {
			room.takenIDs[msg.ParticipantID] = true
		}
		ids := make([]string, 0, len(room.takenIDs))
		for id := range room.takenIDs {
			ids = append(ids, id)
		}
		room.mu.Unlock()
		takenMsg, _ := json.Marshal(map[string]interface{}{"type": "taken-ids", "ids": ids})
		s.broadcast(room, takenMsg, nil)

	case "request-state":
		s.broadcast(room, raw, sender)

	case "vote-stage":
		s.handleVote(room, msg)

	default:
		s.broadcast(room, raw, sender)
	}
}

func (s *Server) handleVote(room *Room, msg incomingMsg) {
	room.mu.Lock()
	defer room.mu.Unlock()

	if msg.Stage == "" || msg.ParticipantID == "" {
		return
	}

	if room.votes[msg.Stage] == nil {
		room.votes[msg.Stage] = make(map[string]bool)
	}
	room.votes[msg.Stage][msg.ParticipantID] = true

	count := len(room.votes[msg.Stage])
	total := len(room.clients)
	threshold := int(math.Ceil(float64(total) * 0.4))

	// Broadcast vote update
	update, _ := json.Marshal(map[string]interface{}{
		"type": "vote-update", "stage": msg.Stage, "count": count, "total": total,
	})
	for client := range room.clients {
		_ = client.writeMessage(update)
	}

	// Check threshold
	if count >= threshold {
		nav, _ := json.Marshal(map[string]interface{}{"type": "navigate-stage", "stage": msg.Stage})
		for client := range room.clients {
			_ = client.writeMessage(nav)
		}
		delete(room.votes, msg.Stage)
	}
}

// StartBackground starts the server on the given port in a goroutine.
// If port is 0, an available port is chosen. Returns the actual port.
func StartBackground(port int) (int, error) {
	srv := New()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return 0, err
	}
	actualPort := ln.Addr().(*net.TCPAddr).Port
	go func() {
		_ = http.Serve(ln, srv.Handler())
	}()
	return actualPort, nil
}
