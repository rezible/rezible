package schematypes

type KnowledgeGraphSubjectState struct {
	DisplayName string         `json:"display_name,omitempty"`
	Description string         `json:"description,omitempty"`
	Properties  map[string]any `json:"properties"`
}
