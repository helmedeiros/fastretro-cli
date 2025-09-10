package domain

import "github.com/helmedeiros/fastretro-cli/internal/protocol"

// FlatActionItem is an action item extracted from a completed retro.
type FlatActionItem struct {
	NoteID      string `json:"noteId"`
	Text        string `json:"text"`
	ParentText  string `json:"parentText"`
	OwnerName   string `json:"ownerName,omitempty"`
	CompletedAt string `json:"completedAt"`
	Done        bool   `json:"done"`
}

// CompletedRetro stores a snapshot of a finished retrospective.
type CompletedRetro struct {
	ID          string              `json:"id"`
	CompletedAt string              `json:"completedAt"`
	ActionItems []FlatActionItem    `json:"actionItems"`
	FullState   protocol.RetroState `json:"fullState"`
}

// RetroHistoryState holds the history of completed retrospectives.
type RetroHistoryState struct {
	Completed []CompletedRetro `json:"completed"`
}

const manualRetroID = "__manual__"

// NewHistory creates an empty history.
func NewHistory() RetroHistoryState {
	return RetroHistoryState{Completed: []CompletedRetro{}}
}

// AddCompletedRetro appends a completed retro to history.
func AddCompletedRetro(h RetroHistoryState, entry CompletedRetro) RetroHistoryState {
	completed := make([]CompletedRetro, len(h.Completed))
	copy(completed, h.Completed)
	return RetroHistoryState{Completed: append(completed, entry)}
}

// GetAllActionItems flattens all action items from all completed retros.
func GetAllActionItems(h RetroHistoryState) []FlatActionItem {
	var items []FlatActionItem
	for _, r := range h.Completed {
		items = append(items, r.ActionItems...)
	}
	return items
}

// GetOpenActionItems returns action items that are not done.
func GetOpenActionItems(h RetroHistoryState) []FlatActionItem {
	var items []FlatActionItem
	for _, item := range GetAllActionItems(h) {
		if !item.Done {
			items = append(items, item)
		}
	}
	return items
}

// ToggleActionItemDone toggles the done status of an action item.
func ToggleActionItemDone(h RetroHistoryState, noteID string) RetroHistoryState {
	return mapActionItem(h, noteID, func(item *FlatActionItem) {
		item.Done = !item.Done
	})
}

// ReassignActionItem changes the owner of an action item.
func ReassignActionItem(h RetroHistoryState, noteID, ownerName string) RetroHistoryState {
	return mapActionItem(h, noteID, func(item *FlatActionItem) {
		item.OwnerName = ownerName
	})
}

// EditActionItemText updates the text of an action item.
func EditActionItemText(h RetroHistoryState, noteID, text string) RetroHistoryState {
	return mapActionItem(h, noteID, func(item *FlatActionItem) {
		item.Text = text
	})
}

// RemoveActionItem deletes an action item by noteID.
func RemoveActionItem(h RetroHistoryState, noteID string) RetroHistoryState {
	completed := make([]CompletedRetro, len(h.Completed))
	for i, r := range h.Completed {
		var items []FlatActionItem
		for _, item := range r.ActionItems {
			if item.NoteID != noteID {
				items = append(items, item)
			}
		}
		if items == nil {
			items = []FlatActionItem{}
		}
		completed[i] = CompletedRetro{
			ID:          r.ID,
			CompletedAt: r.CompletedAt,
			ActionItems: items,
			FullState:   r.FullState,
		}
	}
	return RetroHistoryState{Completed: completed}
}

// AddManualActionItem adds an action item not tied to a retro.
func AddManualActionItem(h RetroHistoryState, item FlatActionItem) RetroHistoryState {
	// Find or create the manual retro entry
	completed := make([]CompletedRetro, len(h.Completed))
	copy(completed, h.Completed)

	for i, r := range completed {
		if r.ID == manualRetroID {
			items := make([]FlatActionItem, len(r.ActionItems))
			copy(items, r.ActionItems)
			completed[i].ActionItems = append(items, item)
			return RetroHistoryState{Completed: completed}
		}
	}
	// No manual entry yet, create one
	manual := CompletedRetro{
		ID:          manualRetroID,
		CompletedAt: item.CompletedAt,
		ActionItems: []FlatActionItem{item},
	}
	return RetroHistoryState{Completed: append(completed, manual)}
}

func mapActionItem(h RetroHistoryState, noteID string, fn func(*FlatActionItem)) RetroHistoryState {
	completed := make([]CompletedRetro, len(h.Completed))
	for i, r := range h.Completed {
		items := make([]FlatActionItem, len(r.ActionItems))
		copy(items, r.ActionItems)
		for j := range items {
			if items[j].NoteID == noteID {
				fn(&items[j])
			}
		}
		completed[i] = CompletedRetro{
			ID:          r.ID,
			CompletedAt: r.CompletedAt,
			ActionItems: items,
			FullState:   r.FullState,
		}
	}
	return RetroHistoryState{Completed: completed}
}
