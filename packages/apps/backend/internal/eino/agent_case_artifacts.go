package eino

import (
	rez "github.com/rezible/rezible"
)

func caseArtifactFromInput(input rez.AgentCaseArtifactInput) AgentCaseArtifact {
	return AgentCaseArtifact{
		Kind:     input.Kind,
		Role:     input.Role,
		Name:     input.Name,
		Payload:  input.Payload,
		Redacted: input.Redacted,
	}
}

func caseArtifactsFromInputs(inputs []rez.AgentCaseArtifactInput) []AgentCaseArtifact {
	artifacts := make([]AgentCaseArtifact, len(inputs))
	for i, input := range inputs {
		artifacts[i] = caseArtifactFromInput(input)
	}
	return artifacts
}
