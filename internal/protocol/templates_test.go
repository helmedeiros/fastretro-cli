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

func TestGetCheckTemplate_HealthCheck(t *testing.T) {
	tmpl := GetCheckTemplate("health-check")
	if tmpl.ID != "health-check" {
		t.Errorf("id: got %q", tmpl.ID)
	}
	if tmpl.Name != "Health Check" {
		t.Errorf("name: got %q", tmpl.Name)
	}
	if len(tmpl.Questions) != 9 {
		t.Errorf("expected 9 questions, got %d", len(tmpl.Questions))
	}
}

func TestGetCheckTemplate_Unknown(t *testing.T) {
	tmpl := GetCheckTemplate("nonexistent")
	if tmpl.ID != "health-check" {
		t.Errorf("expected default 'health-check', got %q", tmpl.ID)
	}
}

func TestCheckTemplate_QuestionsHaveOptions(t *testing.T) {
	tmpl := GetCheckTemplate("health-check")
	for _, q := range tmpl.Questions {
		if q.ID == "" {
			t.Error("question has empty ID")
		}
		if q.Title == "" {
			t.Errorf("question %q has empty title", q.ID)
		}
		if q.Description == "" {
			t.Errorf("question %q has empty description", q.ID)
		}
		if len(q.Options) == 0 {
			t.Errorf("question %q has no options", q.ID)
		}
		for _, opt := range q.Options {
			if opt.Value < 1 {
				t.Errorf("question %q option value %d < 1", q.ID, opt.Value)
			}
			if opt.Label == "" {
				t.Errorf("question %q option with value %d has empty label", q.ID, opt.Value)
			}
		}
	}
}

func TestCheckTemplate_HealthCheckQuestionTitles(t *testing.T) {
	tmpl := GetCheckTemplate("health-check")
	expected := []string{
		"Ownership", "Value", "Goal Alignment", "Communication",
		"Team Roles", "Velocity", "Support And Resources", "Process", "Fun",
	}
	if len(tmpl.Questions) != len(expected) {
		t.Fatalf("expected %d questions, got %d", len(expected), len(tmpl.Questions))
	}
	for i, q := range tmpl.Questions {
		if q.Title != expected[i] {
			t.Errorf("question %d title: got %q, want %q", i, q.Title, expected[i])
		}
	}
}

func TestNumericOptions(t *testing.T) {
	opts := numericOptions(5)
	if len(opts) != 5 {
		t.Fatalf("expected 5 options, got %d", len(opts))
	}
	for i, opt := range opts {
		if opt.Value != i+1 {
			t.Errorf("option %d value: got %d", i, opt.Value)
		}
	}
}
