package types

import "time"

type OncallEvent struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Kind        string    `json:"kind"`
	Title       string    `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Source      *string   `json:"source,omitempty"`
}
