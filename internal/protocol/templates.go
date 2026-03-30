package protocol

import "fmt"

// ColumnTemplate describes a column within a facilitation template.
type ColumnTemplate struct {
	ID          string
	Title       string
	Description string
}

// FacilitationTemplate describes a retrospective format with its columns.
type FacilitationTemplate struct {
	ID      string
	Name    string
	Columns []ColumnTemplate
}

// Templates mirrors the web app's facilitation templates.
var Templates = []FacilitationTemplate{
	{
		ID:   "start-stop",
		Name: "Start / Stop",
		Columns: []ColumnTemplate{
			{ID: "stop", Title: "Stop", Description: "What factors are slowing us down or holding us back?"},
			{ID: "start", Title: "Start", Description: "What factors are driving us forward and enabling our success?"},
		},
	},
	{
		ID:   "anchors-engines",
		Name: "Anchors & Engines",
		Columns: []ColumnTemplate{
			{ID: "anchors", Title: "Anchors", Description: "What factors are slowing us down or holding us back?"},
			{ID: "engines", Title: "Engines", Description: "What factors are driving us forward and enabling our success?"},
		},
	},
	{
		ID:   "mad-sad-glad",
		Name: "Mad Sad Glad",
		Columns: []ColumnTemplate{
			{ID: "mad", Title: "Mad", Description: "What made you frustrated or angry?"},
			{ID: "sad", Title: "Sad", Description: "What disappointed you or could have been better?"},
			{ID: "glad", Title: "Glad", Description: "What made you happy or went well?"},
		},
	},
	{
		ID:   "four-ls",
		Name: "Four Ls",
		Columns: []ColumnTemplate{
			{ID: "liked", Title: "Liked", Description: "What did you like about this iteration?"},
			{ID: "learned", Title: "Learned", Description: "What did you learn?"},
			{ID: "lacked", Title: "Lacked", Description: "What was missing or lacking?"},
			{ID: "longed-for", Title: "Longed for", Description: "What do you wish for in the future?"},
		},
	},
	{
		ID:   "kalm",
		Name: "KALM",
		Columns: []ColumnTemplate{
			{ID: "keep", Title: "Keep", Description: "What should we keep doing?"},
			{ID: "add", Title: "Add", Description: "What should we add or start doing?"},
			{ID: "less", Title: "Less", Description: "What should we do less of?"},
			{ID: "more", Title: "More", Description: "What should we do more of?"},
		},
	},
	{
		ID:   "starfish",
		Name: "Starfish",
		Columns: []ColumnTemplate{
			{ID: "start", Title: "Start", Description: "What should we start doing?"},
			{ID: "more-of", Title: "More of", Description: "What should we do more of?"},
			{ID: "continue", Title: "Continue", Description: "What should we continue doing?"},
			{ID: "less-of", Title: "Less of", Description: "What should we do less of?"},
			{ID: "stop", Title: "Stop", Description: "What should we stop doing?"},
		},
	},
}

// GetTemplate returns the template for the given ID, or the first template as default.
func GetTemplate(id string) FacilitationTemplate {
	for _, t := range Templates {
		if t.ID == id {
			return t
		}
	}
	return Templates[0]
}

// AnswerOption represents a selectable answer for a check question.
type AnswerOption struct {
	Value int
	Label string
}

// CheckQuestion describes a single question within a check template.
type CheckQuestion struct {
	ID          string
	Title       string
	Description string
	Options     []AnswerOption
}

// CheckTemplate describes a check format with its questions.
type CheckTemplate struct {
	ID        string
	Name      string
	Questions []CheckQuestion
}

func numericOptions(n int) []AnswerOption {
	opts := make([]AnswerOption, n)
	for i := range opts {
		opts[i] = AnswerOption{Value: i + 1, Label: fmt.Sprintf("%d", i+1)}
	}
	return opts
}

// CheckTemplates mirrors the web app's check templates.
var CheckTemplates = []CheckTemplate{
	{
		ID:   "health-check",
		Name: "Health Check",
		Questions: []CheckQuestion{
			{ID: "ownership", Title: "Ownership", Description: "The team has clear ownership or a dedicated product owner who is accountable for the team's results and champions the mission inside and outside of the team.", Options: numericOptions(5)},
			{ID: "value", Title: "Value", Description: "We can define and measure the value we provide to the business and the user.", Options: numericOptions(5)},
			{ID: "goal-alignment", Title: "Goal Alignment", Description: "Everyone understands why they are here, supports the idea, and believes they have what it takes to create solutions that add value.", Options: numericOptions(5)},
			{ID: "communication", Title: "Communication", Description: "We have clear and consistent communication that ensures that issues are shared, conflict is reduced, and everyone can work with greater efficiency.", Options: numericOptions(5)},
			{ID: "team-roles", Title: "Team Roles", Description: "The current team skill set is right for the current stage and there are clear roles and responsibilities for each person in the team.", Options: numericOptions(5)},
			{ID: "velocity", Title: "Velocity", Description: "We learn and implement lessons leading to incremental progress in iterations and production as we go.", Options: numericOptions(5)},
			{ID: "support-and-resources", Title: "Support And Resources", Description: "We are equipped with the right tools and resources and can easily access support from within and outside the team.", Options: numericOptions(5)},
			{ID: "process", Title: "Process", Description: "Our processes are aligned, effective, and free of unnecessary delays and blocks. We have metrics in place to measure our goals.", Options: numericOptions(5)},
			{ID: "fun", Title: "Fun", Description: "We enjoy our work and working as a team. We are being challenged and can develop our skill set or acquire new ones.", Options: numericOptions(5)},
		},
	},
}

// GetCheckTemplate returns the check template for the given ID, or the first template as default.
func GetCheckTemplate(id string) CheckTemplate {
	for _, t := range CheckTemplates {
		if t.ID == id {
			return t
		}
	}
	return CheckTemplates[0]
}

// GetColumnTemplate returns the column template for the given template and column IDs.
func GetColumnTemplate(templateID, columnID string) (ColumnTemplate, bool) {
	tmpl := GetTemplate(templateID)
	for _, col := range tmpl.Columns {
		if col.ID == columnID {
			return col, true
		}
	}
	return ColumnTemplate{}, false
}
