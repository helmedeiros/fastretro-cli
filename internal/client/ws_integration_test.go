package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func startTestServer(t *testing.T, handler func(*websocket.Conn)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)
			return
		}
		defer conn.Close()
		handler(conn)
	}))
	return srv
}

func TestConnect_Success(t *testing.T) {
	srv := startTestServer(t, func(conn *websocket.Conn) {
		// Just accept and hold the connection
		conn.ReadMessage()
	})
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "", 1)
	c, err := Connect("TESTROOM", "http://"+wsURL)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer c.Close()

	if c.RoomCode != "TESTROOM" {
		t.Errorf("room code: got %q, want 'TESTROOM'", c.RoomCode)
	}
}

func TestConnect_InvalidCode(t *testing.T) {
	_, err := Connect("AB", "http://localhost:0")
	if err == nil {
		t.Error("expected error for invalid room code")
	}
}

func TestConnect_ServerDown(t *testing.T) {
	_, err := Connect("TESTROOM", "http://localhost:1")
	if err == nil {
		t.Error("expected error when server is down")
	}
}

func TestReadMessage(t *testing.T) {
	srv := startTestServer(t, func(conn *websocket.Conn) {
		msg := map[string]interface{}{
			"type":  "peer-count",
			"count": 3,
		}
		data, _ := json.Marshal(msg)
		conn.WriteMessage(websocket.TextMessage, data)
		conn.ReadMessage() // keep alive
	})
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "", 1)
	c, err := Connect("TESTROOM", "http://"+wsURL)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer c.Close()

	inmsg, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if inmsg.Type != "peer-count" {
		t.Errorf("type: got %q, want 'peer-count'", inmsg.Type)
	}
	if inmsg.Count != 3 {
		t.Errorf("count: got %d, want 3", inmsg.Count)
	}
}

func TestSend(t *testing.T) {
	received := make(chan string, 1)
	srv := startTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err == nil {
			received <- string(data)
		}
	})
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "", 1)
	c, err := Connect("TESTROOM", "http://"+wsURL)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer c.Close()

	err = c.Send([]byte(`{"type":"test"}`))
	if err != nil {
		t.Fatalf("send error: %v", err)
	}

	msg := <-received
	if !strings.Contains(msg, "test") {
		t.Errorf("expected 'test' in message, got %q", msg)
	}
}

func TestClaimIdentity(t *testing.T) {
	received := make(chan string, 1)
	srv := startTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err == nil {
			received <- string(data)
		}
	})
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "", 1)
	c, err := Connect("TESTROOM", "http://"+wsURL)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer c.Close()

	err = c.ClaimIdentity("alice")
	if err != nil {
		t.Fatalf("claim error: %v", err)
	}

	msg := <-received
	var result map[string]string
	json.Unmarshal([]byte(msg), &result)
	if result["type"] != "claim-identity" {
		t.Errorf("type: got %q", result["type"])
	}
	if result["participantId"] != "alice" {
		t.Errorf("participantId: got %q", result["participantId"])
	}
}

func TestVoteStage(t *testing.T) {
	received := make(chan string, 1)
	srv := startTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err == nil {
			received <- string(data)
		}
	})
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "", 1)
	c, err := Connect("TESTROOM", "http://"+wsURL)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer c.Close()

	err = c.VoteStage("discuss", "alice")
	if err != nil {
		t.Fatalf("vote stage error: %v", err)
	}

	msg := <-received
	var result map[string]string
	json.Unmarshal([]byte(msg), &result)
	if result["type"] != "vote-stage" {
		t.Errorf("type: got %q", result["type"])
	}
	if result["stage"] != "discuss" {
		t.Errorf("stage: got %q", result["stage"])
	}
}

func TestSendState(t *testing.T) {
	received := make(chan string, 1)
	srv := startTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err == nil {
			received <- string(data)
		}
	})
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "", 1)
	c, err := Connect("TESTROOM", "http://"+wsURL)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer c.Close()

	state := &protocol.RetroState{
		Stage: "brainstorm",
		Meta:  protocol.RetroMeta{Name: "Test"},
	}
	err = c.SendState(state)
	if err != nil {
		t.Fatalf("send state error: %v", err)
	}

	msg := <-received
	if !strings.Contains(msg, "brainstorm") {
		t.Errorf("expected 'brainstorm' in message, got %q", msg)
	}
}

func TestRequestState(t *testing.T) {
	received := make(chan string, 1)
	srv := startTestServer(t, func(conn *websocket.Conn) {
		_, data, err := conn.ReadMessage()
		if err == nil {
			received <- string(data)
		}
	})
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "", 1)
	c, err := Connect("TESTROOM", "http://"+wsURL)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer c.Close()

	err = c.RequestState()
	if err != nil {
		t.Fatalf("request state error: %v", err)
	}

	msg := <-received
	if !strings.Contains(msg, "request-state") {
		t.Errorf("expected 'request-state' in message, got %q", msg)
	}
}

func TestClose(t *testing.T) {
	srv := startTestServer(t, func(conn *websocket.Conn) {
		conn.ReadMessage()
	})
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "", 1)
	c, err := Connect("TESTROOM", "http://"+wsURL)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}

	err = c.Close()
	if err != nil {
		t.Errorf("close error: %v", err)
	}
}

func TestClose_NilConn(t *testing.T) {
	c := &Client{}
	err := c.Close()
	if err != nil {
		t.Errorf("close nil conn should not error: %v", err)
	}
}

func TestReadMessage_ClosedConnection(t *testing.T) {
	srv := startTestServer(t, func(conn *websocket.Conn) {
		conn.Close()
	})
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "", 1)
	c, err := Connect("TESTROOM", "http://"+wsURL)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer c.Close()

	_, err = c.ReadMessage()
	if err == nil {
		t.Error("expected error reading from closed connection")
	}
}
