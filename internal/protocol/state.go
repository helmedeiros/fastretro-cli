package protocol

// RetroState represents the full state of a retrospective session.
type RetroState struct {
	Stage           string            `json:"stage"`
	Meta            RetroMeta         `json:"meta"`
	Participants    []Participant     `json:"participants"`
	Timer           *Timer            `json:"timer"`
	Icebreaker      *Icebreaker       `json:"icebreaker"`
	Cards           []Card            `json:"cards"`
	Groups          []Group           `json:"groups"`
	Votes           []Vote            `json:"votes"`
	VoteBudget      int               `json:"voteBudget"`
	Discuss         *DiscussState     `json:"discuss"`
	DiscussNotes    []DiscussNote     `json:"discussNotes"`
	ActionOwners    map[string]string `json:"actionItemOwners"`
	SurveyResponses []SurveyResponse  `json:"surveyResponses"`
}

// SurveyResponse represents a participant's rating for a check question.
type SurveyResponse struct {
	ID            string `json:"id"`
	ParticipantID string `json:"participantId"`
	QuestionID    string `json:"questionId"`
	Rating        int    `json:"rating"`
	Comment       string `json:"comment"`
}

type RetroMeta struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Date       string `json:"date"`
	Context    string `json:"context"`
	TemplateID string `json:"templateId"`
}

type Participant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Timer struct {
	Status      string `json:"status"`
	DurationMs  int    `json:"durationMs"`
	ElapsedMs   int    `json:"elapsedMs"`
	RemainingMs int    `json:"remainingMs"`
}

type Icebreaker struct {
	Question       string   `json:"question"`
	Questions      []string `json:"questions"`
	ParticipantIDs []string `json:"participantIds"`
	CurrentIndex   int      `json:"currentIndex"`
}

type Card struct {
	ID       string `json:"id"`
	ColumnID string `json:"columnId"`
	Text     string `json:"text"`
}

type Group struct {
	ID       string   `json:"id"`
	ColumnID string   `json:"columnId"`
	Name     string   `json:"name"`
	CardIDs  []string `json:"cardIds"`
}

type Vote struct {
	ParticipantID string `json:"participantId"`
	CardID        string `json:"cardId"`
}

type DiscussState struct {
	Order        []string `json:"order"`
	CurrentIndex int      `json:"currentIndex"`
	Segment      string   `json:"segment"`
}

type DiscussNote struct {
	ID           string `json:"id"`
	ParentCardID string `json:"parentCardId"`
	Lane         string `json:"lane"`
	Text         string `json:"text"`
}
