package mcp

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
)

func extractIdParam(uri string) (uuid.UUID, error) {
	parts := strings.Split(uri, "://")
	if len(parts) < 2 {
		return uuid.Nil, fmt.Errorf("mcp: invalid URI: %s", uri)
	}
	return uuid.Parse(parts[1])
}
