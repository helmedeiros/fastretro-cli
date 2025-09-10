package domain

import "strings"

// TeamEntry represents a team in the global registry.
type TeamEntry struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

// AddTeamEntry adds a new entry to the registry. Returns error if name is empty or duplicate.
func AddTeamEntry(entries []TeamEntry, id, name, createdAt string) ([]TeamEntry, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return entries, ErrEmptyName
	}
	for _, e := range entries {
		if strings.EqualFold(e.Name, name) {
			return entries, ErrDuplicateName
		}
	}
	result := make([]TeamEntry, len(entries))
	copy(result, entries)
	return append(result, TeamEntry{ID: id, Name: name, CreatedAt: createdAt}), nil
}

// RemoveTeamEntry removes an entry by ID.
func RemoveTeamEntry(entries []TeamEntry, id string) []TeamEntry {
	var result []TeamEntry
	for _, e := range entries {
		if e.ID != id {
			result = append(result, e)
		}
	}
	if result == nil {
		return []TeamEntry{}
	}
	return result
}

// RenameTeamEntry renames a team entry. Returns error if name is empty or duplicate.
func RenameTeamEntry(entries []TeamEntry, id, name string) ([]TeamEntry, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return entries, ErrEmptyName
	}
	for _, e := range entries {
		if e.ID != id && strings.EqualFold(e.Name, name) {
			return entries, ErrDuplicateName
		}
	}
	result := make([]TeamEntry, len(entries))
	copy(result, entries)
	for i, e := range result {
		if e.ID == id {
			result[i].Name = name
			return result, nil
		}
	}
	return entries, ErrMemberNotFound // reuse: "not found"
}

// FindTeamEntryByName finds a team by name (case-insensitive).
func FindTeamEntryByName(entries []TeamEntry, name string) (TeamEntry, bool) {
	for _, e := range entries {
		if strings.EqualFold(e.Name, name) {
			return e, true
		}
	}
	return TeamEntry{}, false
}
