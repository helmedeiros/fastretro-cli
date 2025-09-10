package domain

import (
	"errors"
	"strings"
)

// TeamMember represents a member of a team.
type TeamMember struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Agreement represents a team commitment.
type Agreement struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}

// TeamState holds the members and agreements for a team.
type TeamState struct {
	Members    []TeamMember `json:"members"`
	Agreements []Agreement  `json:"agreements"`
}

var (
	ErrEmptyName      = errors.New("name must not be empty")
	ErrDuplicateName  = errors.New("name already exists")
	ErrMemberNotFound = errors.New("member not found")
	ErrAgreementNotFound = errors.New("agreement not found")
)

// NewTeam creates an empty team state.
func NewTeam() TeamState {
	return TeamState{
		Members:    []TeamMember{},
		Agreements: []Agreement{},
	}
}

// AddMember adds a member to the team. Returns error if name is empty or duplicate.
func AddMember(state TeamState, id, name string) (TeamState, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return state, ErrEmptyName
	}
	for _, m := range state.Members {
		if strings.EqualFold(m.Name, name) {
			return state, ErrDuplicateName
		}
	}
	result := TeamState{
		Members:    append(copyMembers(state.Members), TeamMember{ID: id, Name: name}),
		Agreements: state.Agreements,
	}
	return result, nil
}

// RemoveMember removes a member by ID.
func RemoveMember(state TeamState, id string) (TeamState, error) {
	found := false
	var members []TeamMember
	for _, m := range state.Members {
		if m.ID == id {
			found = true
			continue
		}
		members = append(members, m)
	}
	if !found {
		return state, ErrMemberNotFound
	}
	if members == nil {
		members = []TeamMember{}
	}
	return TeamState{Members: members, Agreements: state.Agreements}, nil
}

// AddAgreement adds a new agreement.
func AddAgreement(state TeamState, id, text, createdAt string) (TeamState, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return state, ErrEmptyName
	}
	result := TeamState{
		Members:    state.Members,
		Agreements: append(copyAgreements(state.Agreements), Agreement{ID: id, Text: text, CreatedAt: createdAt}),
	}
	return result, nil
}

// EditAgreement updates agreement text by ID.
func EditAgreement(state TeamState, id, text string) (TeamState, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return state, ErrEmptyName
	}
	agreements := copyAgreements(state.Agreements)
	found := false
	for i, a := range agreements {
		if a.ID == id {
			agreements[i].Text = text
			found = true
			break
		}
	}
	if !found {
		return state, ErrAgreementNotFound
	}
	return TeamState{Members: state.Members, Agreements: agreements}, nil
}

// RemoveAgreement removes an agreement by ID.
func RemoveAgreement(state TeamState, id string) TeamState {
	var agreements []Agreement
	for _, a := range state.Agreements {
		if a.ID != id {
			agreements = append(agreements, a)
		}
	}
	if agreements == nil {
		agreements = []Agreement{}
	}
	return TeamState{Members: state.Members, Agreements: agreements}
}

func copyMembers(m []TeamMember) []TeamMember {
	c := make([]TeamMember, len(m))
	copy(c, m)
	return c
}

func copyAgreements(a []Agreement) []Agreement {
	c := make([]Agreement, len(a))
	copy(c, a)
	return c
}
