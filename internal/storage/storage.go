package storage

import "github.com/helmedeiros/fastretro-cli/internal/domain"

// TeamRepository manages persistent storage for a single team's data.
type TeamRepository interface {
	LoadTeam() (domain.TeamState, error)
	SaveTeam(state domain.TeamState) error
	LoadHistory() (domain.RetroHistoryState, error)
	SaveHistory(state domain.RetroHistoryState) error
}

// RegistryRepository manages the global team registry and config.
type RegistryRepository interface {
	List() ([]domain.TeamEntry, error)
	Save(entries []domain.TeamEntry) error
	SelectedTeamID() (string, error)
	SetSelectedTeamID(id string) error
}
