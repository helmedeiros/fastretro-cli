package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/helmedeiros/fastretro-cli/internal/domain"
)

// JSONRegistryRepo implements RegistryRepository using JSON files.
type JSONRegistryRepo struct {
	baseDir string
}

// NewJSONRegistryRepo creates a registry repo rooted at baseDir.
func NewJSONRegistryRepo(baseDir string) *JSONRegistryRepo {
	return &JSONRegistryRepo{baseDir: baseDir}
}

type configFile struct {
	SelectedTeamID string `json:"selectedTeamId"`
	DefaultMember  string `json:"defaultMember,omitempty"`
}

func (r *JSONRegistryRepo) List() ([]domain.TeamEntry, error) {
	path := filepath.Join(r.baseDir, "teams", "registry.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.TeamEntry{}, nil
		}
		return nil, err
	}
	var entries []domain.TeamEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	if entries == nil {
		return []domain.TeamEntry{}, nil
	}
	return entries, nil
}

func (r *JSONRegistryRepo) Save(entries []domain.TeamEntry) error {
	dir := filepath.Join(r.baseDir, "teams")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return atomicWrite(filepath.Join(dir, "registry.json"), entries)
}

func (r *JSONRegistryRepo) SelectedTeamID() (string, error) {
	path := filepath.Join(r.baseDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var cfg configFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	return cfg.SelectedTeamID, nil
}

func (r *JSONRegistryRepo) SetSelectedTeamID(id string) error {
	if err := os.MkdirAll(r.baseDir, 0755); err != nil {
		return err
	}
	return atomicWrite(filepath.Join(r.baseDir, "config.json"), configFile{SelectedTeamID: id})
}

// TeamDir returns the directory path for a specific team.
func (r *JSONRegistryRepo) TeamDir(teamID string) string {
	return filepath.Join(r.baseDir, "teams", teamID)
}

type identityFile struct {
	RoomCode      string `json:"roomCode"`
	ParticipantID string `json:"participantId"`
}

// LoadIdentity returns the persisted identity for the given room code, if any.
func (r *JSONRegistryRepo) LoadIdentity(roomCode string) string {
	path := filepath.Join(r.baseDir, "identity.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var id identityFile
	if err := json.Unmarshal(data, &id); err != nil {
		return ""
	}
	if id.RoomCode != roomCode {
		return ""
	}
	return id.ParticipantID
}

// LoadDefaultMember returns the stored default member name, if any.
func (r *JSONRegistryRepo) LoadDefaultMember() string {
	path := filepath.Join(r.baseDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var cfg configFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ""
	}
	return cfg.DefaultMember
}

// SaveDefaultMember persists the default member name.
func (r *JSONRegistryRepo) SaveDefaultMember(name string) {
	path := filepath.Join(r.baseDir, "config.json")
	var cfg configFile
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &cfg)
	}
	cfg.DefaultMember = name
	if err := os.MkdirAll(r.baseDir, 0755); err != nil {
		return
	}
	atomicWrite(path, cfg)
}

// SaveIdentity persists the chosen identity for a room code.
func (r *JSONRegistryRepo) SaveIdentity(roomCode, participantID string) {
	if err := os.MkdirAll(r.baseDir, 0755); err != nil {
		return
	}
	atomicWrite(filepath.Join(r.baseDir, "identity.json"), identityFile{
		RoomCode:      roomCode,
		ParticipantID: participantID,
	})
}
