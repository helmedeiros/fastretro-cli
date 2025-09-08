package protocol

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
