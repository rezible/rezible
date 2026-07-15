package ai

import "github.com/google/uuid"

type ToolDefinition[I any, O any] struct {
	name        string
	description string
}

func (d ToolDefinition[I, O]) Name() string {
	return d.name
}

func (d ToolDefinition[I, O]) Description() string {
	return d.description
}

func defineTool[D ToolDefinition[I, O], I any, O any](name, description string) D {
	return D{name: name, description: description}
}

type (
	WriteAgentOutputArtifactToolOutput struct {
		Status string `json:"status"`
	}

	AgentOutputWithWriteArtifactTool interface {
		WriteArtifactToolDefinition() (name string, description string)
	}

	WriteAgentOutputArtifactToolDefinition[O AgentOutput] ToolDefinition[O, WriteAgentOutputArtifactToolOutput]
)

type (
	SendChatMessageToolInput struct {
		Message string `json:"message"`
	}

	SendChatMessageToolOutput struct {
		Status string `json:"status"`
	}

	SendChatMessageToolDefinition = ToolDefinition[SendChatMessageToolInput, SendChatMessageToolOutput]

	EventSendChatMessageToolInvoked struct {
		AgentRunId uuid.UUID                `json:"agent_run_id"`
		Input      SendChatMessageToolInput `json:"input"`
	}
)

var SendChatMessageTool = SendChatMessageToolDefinition{
	name:        "send_message",
	description: "send a chat message reply",
}
