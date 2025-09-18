package protocol

import "encoding/json"

// Message types for WebSocket communication.

// SyncTeamInfo carries team metadata over WebSocket.
type SyncTeamInfo struct {
	TeamName   string            `json:"teamName"`
	Members    []TeamInfoMember  `json:"members"`
	Agreements []TeamInfoAgreement `json:"agreements"`
}

// TeamInfoMember is a member entry in team-info messages.
type TeamInfoMember struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TeamInfoAgreement is an agreement entry in team-info messages.
type TeamInfoAgreement struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// IncomingMessage represents a message received from the server.
type IncomingMessage struct {
	Type     string         `json:"type"`
	State    *RetroState    `json:"state,omitempty"`
	Stage    string         `json:"stage,omitempty"`
	Count    int            `json:"count,omitempty"`
	Total    int            `json:"total,omitempty"`
	IDs      []string       `json:"ids,omitempty"`
	PID      string         `json:"participantId,omitempty"`
	TeamInfo *SyncTeamInfo  `json:"teamInfo,omitempty"`
}

// ParseMessage parses a raw JSON message from the WebSocket.
func ParseMessage(data []byte) (IncomingMessage, error) {
	var msg IncomingMessage
	err := json.Unmarshal(data, &msg)
	return msg, err
}

// --- Outgoing messages ---

// StateMessage broadcasts the full retro state.
func StateMessage(state *RetroState) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":  "state",
		"state": state,
	})
}

// VoteStageMessage sends a stage navigation vote.
func VoteStageMessage(stage, participantID string) ([]byte, error) {
	return json.Marshal(map[string]string{
		"type":          "vote-stage",
		"stage":         stage,
		"participantId": participantID,
	})
}

// ClaimIdentityMessage claims a participant identity.
func ClaimIdentityMessage(participantID string) ([]byte, error) {
	return json.Marshal(map[string]string{
		"type":          "claim-identity",
		"participantId": participantID,
	})
}

// RequestStateMessage asks other clients to broadcast their state.
func RequestStateMessage() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type": "request-state",
	})
}
