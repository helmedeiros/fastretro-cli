package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

// Client manages the WebSocket connection to a fastRetro room.
type Client struct {
	conn     *websocket.Conn
	RoomCode string
}

// Connect establishes a WebSocket connection to the given room.
// input can be a room code like "ABC-123-DEF" or a URL like "http://host:port/#room=CODE"
func Connect(input string, serverURL string) (*Client, error) {
	code := extractRoomCode(input)
	if code == "" {
		return nil, fmt.Errorf("invalid room code or URL: %s", input)
	}

	wsURL := fmt.Sprintf("%s/__ws/room/%s", toWSURL(serverURL), code)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to room %s: %w", code, err)
	}

	return &Client{conn: conn, RoomCode: code}, nil
}

// ReadMessage reads the next message from the server.
func (c *Client) ReadMessage() (protocol.IncomingMessage, error) {
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return protocol.IncomingMessage{}, err
	}
	return protocol.ParseMessage(data)
}

// Send sends raw bytes to the server.
func (c *Client) Send(data []byte) error {
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// ClaimIdentity sends a claim-identity message.
func (c *Client) ClaimIdentity(participantID string) error {
	msg, err := protocol.ClaimIdentityMessage(participantID)
	if err != nil {
		return err
	}
	return c.Send(msg)
}

// VoteStage sends a vote-stage message.
func (c *Client) VoteStage(stage, participantID string) error {
	msg, err := protocol.VoteStageMessage(stage, participantID)
	if err != nil {
		return err
	}
	return c.Send(msg)
}

// SendState broadcasts the full retro state.
func (c *Client) SendState(state *protocol.RetroState) error {
	msg, err := protocol.StateMessage(state)
	if err != nil {
		return err
	}
	return c.Send(msg)
}

// RequestState asks other clients to send their state.
func (c *Client) RequestState() error {
	msg, err := protocol.RequestStateMessage()
	if err != nil {
		return err
	}
	return c.Send(msg)
}

// ShareURL returns a shareable URL for the room.
func (c *Client) ShareURL(serverURL string) string {
	return fmt.Sprintf("%s/#room=%s", strings.TrimRight(serverURL, "/"), c.RoomCode)
}

// CreateRoom calls the server API to create a new room and returns the code.
func CreateRoom(serverURL string) (string, error) {
	url := fmt.Sprintf("%s/__api/rooms", strings.TrimRight(serverURL, "/"))
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create room: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse room response: %w", err)
	}
	return result.Code, nil
}

// Close closes the WebSocket connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func extractRoomCode(input string) string {
	input = strings.TrimSpace(input)

	// Direct room code: ABC-123-DEF
	if isRoomCode(input) {
		return input
	}

	// URL with hash: http://host/#room=CODE
	if strings.Contains(input, "#room=") {
		parts := strings.SplitN(input, "#room=", 2)
		if len(parts) == 2 && isRoomCode(parts[1]) {
			return parts[1]
		}
	}

	// URL path: try to parse
	u, err := url.Parse(input)
	if err == nil && u.Fragment != "" {
		if strings.HasPrefix(u.Fragment, "room=") {
			code := strings.TrimPrefix(u.Fragment, "room=")
			if isRoomCode(code) {
				return code
			}
		}
	}

	return ""
}

func isRoomCode(s string) bool {
	// Room codes are like ABC-123-DEF (alphanumeric with dashes)
	if len(s) < 5 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' ||
			(c >= 'a' && c <= 'z')) {
			return false
		}
	}
	return true
}

func toWSURL(serverURL string) string {
	serverURL = strings.TrimRight(serverURL, "/")
	serverURL = strings.Replace(serverURL, "https://", "wss://", 1)
	serverURL = strings.Replace(serverURL, "http://", "ws://", 1)
	if !strings.HasPrefix(serverURL, "ws") {
		serverURL = "ws://" + serverURL
	}
	return serverURL
}
