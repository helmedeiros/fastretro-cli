package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func testServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	srv := New()
	ts := httptest.NewServer(srv.Handler())
	return srv, ts
}

func wsURL(ts *httptest.Server, code string) string {
	return "ws" + strings.TrimPrefix(ts.URL, "http") + "/__ws/room/" + code
}

func connectWS(t *testing.T, ts *httptest.Server, code string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL(ts, code), nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	return conn
}

// drainMessages reads all available messages within a short window.
func drainMessages(conn *websocket.Conn) {
	for {
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
	conn.SetReadDeadline(time.Time{}) // clear deadline
}

// readMsgOfType reads messages until finding one with the given type.
func readMsgOfType(t *testing.T, conn *websocket.Conn, msgType string) map[string]interface{} {
	t.Helper()
	for i := 0; i < 10; i++ {
		msg := readMsg(t, conn)
		if msg["type"] == msgType {
			return msg
		}
	}
	t.Fatalf("did not receive message of type %q", msgType)
	return nil
}

func readMsg(t *testing.T, conn *websocket.Conn) map[string]interface{} {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	return msg
}

func TestCreateRoom(t *testing.T) {
	_, ts := testServer(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/__api/rooms", "application/json", nil)
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if body["code"] == "" {
		t.Error("expected non-empty room code")
	}
	// Format: XXX-XXX-XXX
	parts := strings.Split(body["code"], "-")
	if len(parts) != 3 {
		t.Errorf("expected 3 code segments, got %d: %q", len(parts), body["code"])
	}
}

func TestCreateRoom_MethodNotAllowed(t *testing.T) {
	_, ts := testServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/__api/rooms")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 405 {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestWebSocket_ConnectAndPeerCount(t *testing.T) {
	_, ts := testServer(t)
	defer ts.Close()

	conn := connectWS(t, ts, "TEST-ROOM")
	defer conn.Close()

	msg := readMsgOfType(t, conn, "peer-count")
	if msg["count"].(float64) != 1 {
		t.Errorf("expected peer count 1, got %v", msg["count"])
	}
}

func TestWebSocket_StateBroadcast(t *testing.T) {
	_, ts := testServer(t)
	defer ts.Close()

	c1 := connectWS(t, ts, "ROOM-A")
	defer c1.Close()
	c2 := connectWS(t, ts, "ROOM-A")
	defer c2.Close()

	// c1 sends state — c2 will get it mixed with initial messages
	time.Sleep(100 * time.Millisecond)
	state := `{"type":"state","state":{"stage":"brainstorm"}}`
	c1.WriteMessage(websocket.TextMessage, []byte(state))

	// c2 reads until it gets a state message with our stage
	msg := readMsgOfType(t, c2, "state")
	stateObj, _ := msg["state"].(map[string]interface{})
	if stateObj == nil || stateObj["stage"] != "brainstorm" {
		t.Errorf("expected brainstorm state, got %v", msg)
	}
}

func TestWebSocket_StatePersistedForNewClient(t *testing.T) {
	_, ts := testServer(t)
	defer ts.Close()

	c1 := connectWS(t, ts, "ROOM-B")
	defer c1.Close()
	drainMessages(c1)

	// Send state
	state := `{"type":"state","state":{"stage":"discuss"}}`
	c1.WriteMessage(websocket.TextMessage, []byte(state))
	time.Sleep(50 * time.Millisecond)

	// New client connects — should receive stored state
	c2 := connectWS(t, ts, "ROOM-B")
	defer c2.Close()

	msg := readMsgOfType(t, c2, "state")
	if msg["type"] != "state" {
		t.Error("expected new client to receive stored state")
	}
}

func TestWebSocket_ClaimIdentity(t *testing.T) {
	_, ts := testServer(t)
	defer ts.Close()

	c1 := connectWS(t, ts, "ROOM-C")
	defer c1.Close()
	time.Sleep(100 * time.Millisecond)

	// c1 claims identity
	claim := `{"type":"claim-identity","participantId":"p1"}`
	c1.WriteMessage(websocket.TextMessage, []byte(claim))

	// Read until we get taken-ids containing p1
	for i := 0; i < 10; i++ {
		msg := readMsg(t, c1)
		if msg["type"] == "taken-ids" {
			ids, _ := msg["ids"].([]interface{})
			for _, id := range ids {
				if id == "p1" {
					return // success
				}
			}
		}
	}
	t.Error("did not receive taken-ids with p1")
}

func TestWebSocket_RoomCleanup(t *testing.T) {
	srv, ts := testServer(t)
	defer ts.Close()

	c1 := connectWS(t, ts, "ROOM-D")
	// Drain
	for i := 0; i < 2; i++ {
		readMsg(t, c1)
	}

	c1.Close()
	time.Sleep(100 * time.Millisecond)

	srv.mu.RLock()
	_, exists := srv.rooms["ROOM-D"]
	srv.mu.RUnlock()

	if exists {
		t.Error("expected room to be deleted after last client disconnects")
	}
}

func TestWebSocket_VoteThreshold(t *testing.T) {
	_, ts := testServer(t)
	defer ts.Close()

	conns := make([]*websocket.Conn, 3)
	for i := range conns {
		conns[i] = connectWS(t, ts, "ROOM-E")
		defer conns[i].Close()
	}
	time.Sleep(200 * time.Millisecond)

	// 40% of 3 = ceil(1.2) = 2 votes needed
	conns[0].WriteMessage(websocket.TextMessage, []byte(`{"type":"vote-stage","stage":"discuss","participantId":"p1"}`))
	time.Sleep(100 * time.Millisecond)
	conns[1].WriteMessage(websocket.TextMessage, []byte(`{"type":"vote-stage","stage":"discuss","participantId":"p2"}`))

	// Client 2 should receive navigate-stage among other messages
	msg := readMsgOfType(t, conns[2], "navigate-stage")
	if msg["stage"] != "discuss" {
		t.Errorf("expected discuss stage, got %v", msg["stage"])
	}
}

func TestStartBackground(t *testing.T) {
	port, err := StartBackground(0)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if port == 0 {
		t.Error("expected non-zero port")
	}

	// Verify server responds
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/__api/rooms", port), "", nil)
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRoomIsolation(t *testing.T) {
	_, ts := testServer(t)
	defer ts.Close()

	c1 := connectWS(t, ts, "ROOM-X")
	defer c1.Close()
	c2 := connectWS(t, ts, "ROOM-Y")
	defer c2.Close()

	// Drain initial
	for i := 0; i < 2; i++ {
		readMsg(t, c1)
	}
	for i := 0; i < 2; i++ {
		readMsg(t, c2)
	}

	// c1 sends to ROOM-X
	c1.WriteMessage(websocket.TextMessage, []byte(`{"type":"state","state":{"stage":"vote"}}`))
	time.Sleep(50 * time.Millisecond)

	// c2 in ROOM-Y should NOT receive it — set short deadline
	c2.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _, err := c2.ReadMessage()
	if err == nil {
		t.Error("expected no message in different room")
	}
}
