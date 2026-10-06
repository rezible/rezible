package schematypes

type KnowledgeGraphSubjectState struct {
	DisplayName string         `json:"display_name,omitempty"`
	Description string         `json:"description,omitempty"`
	Properties  map[string]any `json:"properties"`
}

// IsEmpty reports whether the state has no display name, description or properties.
func (s KnowledgeGraphSubjectState) IsEmpty() bool {
	return s.DisplayName == "" && s.Description == "" && len(s.Properties) == 0
}

// MergeFrom returns the state with each non-empty field of other replacing its value. Empty fields of other
// never clear a value.
func (s KnowledgeGraphSubjectState) MergeFrom(other KnowledgeGraphSubjectState) KnowledgeGraphSubjectState {
	if other.DisplayName != "" {
		s.DisplayName = other.DisplayName
	}
	if other.Description != "" {
		s.Description = other.Description
	}
	if len(other.Properties) > 0 {
		s.Properties = other.Properties
	}
	return s
}
