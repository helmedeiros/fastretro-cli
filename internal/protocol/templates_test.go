package protocol

import "testing"

func TestGetTemplate_Known(t *testing.T) {
	tests := []struct {
		id      string
		name    string
		numCols int
	}{
		{"start-stop", "Start / Stop", 2},
		{"anchors-engines", "Anchors & Engines", 2},
		{"mad-sad-glad", "Mad Sad Glad", 3},
		{"four-ls", "Four Ls", 4},
		{"kalm", "KALM", 4},
		{"starfish", "Starfish", 5},
	}

	for _, tt := range tests {
		tmpl := GetTemplate(tt.id)
		if tmpl.Name != tt.name {
			t.Errorf("GetTemplate(%q).Name = %q, want %q", tt.id, tmpl.Name, tt.name)
		}
		if len(tmpl.Columns) != tt.numCols {
			t.Errorf("GetTemplate(%q) has %d columns, want %d", tt.id, len(tmpl.Columns), tt.numCols)
		}
	}
}

func TestGetTemplate_Unknown(t *testing.T) {
	tmpl := GetTemplate("nonexistent")
	if tmpl.ID != "start-stop" {
		t.Errorf("expected default 'start-stop', got %q", tmpl.ID)
	}
}

func TestGetTemplate_Empty(t *testing.T) {
	tmpl := GetTemplate("")
	if tmpl.ID != "start-stop" {
		t.Errorf("expected default, got %q", tmpl.ID)
	}
}

func TestGetColumnTemplate_Found(t *testing.T) {
	col, ok := GetColumnTemplate("mad-sad-glad", "mad")
	if !ok {
		t.Fatal("expected to find 'mad' column")
	}
	if col.Title != "Mad" {
		t.Errorf("title: got %q", col.Title)
	}
	if col.Description == "" {
		t.Error("expected description")
	}
}

func TestGetColumnTemplate_NotFound(t *testing.T) {
	_, ok := GetColumnTemplate("start-stop", "nonexistent")
	if ok {
		t.Error("should not find nonexistent column")
	}
}

func TestGetColumnTemplate_UnknownTemplate(t *testing.T) {
	col, ok := GetColumnTemplate("unknown", "stop")
	if !ok {
		t.Fatal("should fall back to default template and find 'stop'")
	}
	if col.Title != "Stop" {
		t.Errorf("title: got %q", col.Title)
	}
}

func TestTemplates_AllHaveIDs(t *testing.T) {
	for _, tmpl := range Templates {
		if tmpl.ID == "" {
			t.Error("template has empty ID")
		}
		if tmpl.Name == "" {
			t.Errorf("template %q has empty name", tmpl.ID)
		}
		for _, col := range tmpl.Columns {
			if col.ID == "" {
				t.Errorf("template %q has column with empty ID", tmpl.ID)
			}
			if col.Title == "" {
				t.Errorf("template %q column %q has empty title", tmpl.ID, col.ID)
			}
			if col.Description == "" {
				t.Errorf("template %q column %q has empty description", tmpl.ID, col.ID)
			}
		}
	}
}

func TestTemplates_Count(t *testing.T) {
	if len(Templates) != 6 {
		t.Errorf("expected 6 templates, got %d", len(Templates))
	}
}
