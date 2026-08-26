import type {
	AgentArtifact,
	AgentArtifactChunk,
	AgentMessage,
	AgentMessagePart,
	AgentModelChunk,
	AgentTurn,
} from "$lib/api";

export type TranscriptGroup = { turn?: AgentTurn; messages: AgentMessage[] };

export const groupTranscript = (turns: AgentTurn[], messages: AgentMessage[]) => {
	const sortedTurns = [...turns].sort((a, b) => a.attributes.sequence - b.attributes.sequence);
	const byTurn = new Map(sortedTurns.map((turn) => [turn.id, { turn, messages: [] } as TranscriptGroup]));
	const ungrouped: AgentMessage[] = [];
	for (const message of [...messages].sort((a, b) => a.attributes.sequence - b.attributes.sequence)) {
		const group = byTurn.get(message.attributes.agentTurnId);
		if (group) group.messages.push(message);
		else ungrouped.push(message);
	}
	return { groups: sortedTurns.map((turn) => byTurn.get(turn.id)!), ungrouped };
};

type LiveOverlayModel = {
	turnId: string;
	index: number;
	role: string;
	parts: AgentMessagePart[];
};

export type LiveOverlay = {
	models: Map<string, LiveOverlayModel>;
	artifacts: Map<string, AgentArtifactChunk>;
	reconcilingTurns: Set<string>;
};

export const emptyOverlay = (): LiveOverlay => ({
	models: new Map(),
	artifacts: new Map(),
	reconcilingTurns: new Set(),
});

export const eventMatchesSession = (openSessionId: string, event: { sessionId: string }) =>
	event.sessionId === openSessionId;

export const applyModelChunk = (overlay: LiveOverlay, turnId: string, chunk: AgentModelChunk) => {
	const updatedOverlay = { ...overlay, models: new Map(overlay.models) };
	const key = `${turnId}:${chunk.index}`;
	const previous = updatedOverlay.models.get(key);
	updatedOverlay.models.set(key, {
		turnId,
		index: chunk.index,
		role: chunk.role,
		parts: chunk.aggregated ? chunk.parts : [...(previous?.parts ?? []), ...chunk.parts],
	});
	return updatedOverlay;
};

export const applyArtifactChunk = (overlay: LiveOverlay, chunk: AgentArtifactChunk) => ({
	...overlay,
	artifacts: new Map(overlay.artifacts).set(chunk.name, chunk),
});

export const reconcileTurn = (overlay: LiveOverlay, turnId: string) => {
	const models = new Map([...overlay.models].filter(([, model]) => model.turnId !== turnId));
	const reconcilingTurns = new Set(overlay.reconcilingTurns);
	reconcilingTurns.delete(turnId);
	return { models, artifacts: new Map(), reconcilingTurns };
};

export const mergedArtifacts = (persisted: AgentArtifact[], overlay: LiveOverlay) => {
	const result = new Map(persisted.map((artifact) => [artifact.attributes.name, artifact]));
	for (const [name, artifact] of overlay.artifacts) {
		result.set(name, {
			id: `live:${name}`,
			attributes: { name, parts: artifact.parts, createdAt: "", updatedAt: "" },
		});
	}
	return [...result.values()].sort((a, b) => a.attributes.name.localeCompare(b.attributes.name));
};
