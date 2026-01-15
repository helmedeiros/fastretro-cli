package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/helmedeiros/fastretro-cli/internal/domain"
)

// JSONTeamRepo implements TeamRepository using JSON files.
type JSONTeamRepo struct {
	teamDir string
}

// NewJSONTeamRepo creates a repo for the given team directory.
func NewJSONTeamRepo(teamDir string) *JSONTeamRepo {
	return &JSONTeamRepo{teamDir: teamDir}
}

func (r *JSONTeamRepo) LoadTeam() (domain.TeamState, error) {
	path := filepath.Join(r.teamDir, "team.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.NewTeam(), nil
		}
		return domain.TeamState{}, err
	}
	var state domain.TeamState
	if err := json.Unmarshal(data, &state); err != nil {
		return domain.TeamState{}, err
	}
	if state.Members == nil {
		state.Members = []domain.TeamMember{}
	}
	if state.Agreements == nil {
		state.Agreements = []domain.Agreement{}
	}
	return state, nil
}

func (r *JSONTeamRepo) SaveTeam(state domain.TeamState) error {
	if err := os.MkdirAll(r.teamDir, 0755); err != nil {
		return err
	}
	return atomicWrite(filepath.Join(r.teamDir, "team.json"), state)
}

func (r *JSONTeamRepo) LoadHistory() (domain.RetroHistoryState, error) {
	path := filepath.Join(r.teamDir, "history.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.NewHistory(), nil
		}
		return domain.RetroHistoryState{}, err
	}
	var state domain.RetroHistoryState
	if err := json.Unmarshal(data, &state); err != nil {
		return domain.RetroHistoryState{}, err
	}
	if state.Completed == nil {
		state.Completed = []domain.CompletedRetro{}
	}
	return state, nil
}

func (r *JSONTeamRepo) SaveHistory(state domain.RetroHistoryState) error {
	if err := os.MkdirAll(r.teamDir, 0755); err != nil {
		return err
	}
	return atomicWrite(filepath.Join(r.teamDir, "history.json"), state)
}

func atomicWrite(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) // clean up temp file on rename failure
		return err
	}
	return nil
}
