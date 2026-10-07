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
	artifacts: Map<string, AgentArtifactChunk & { turnId: string }>;
	attempts: Map<string, { startedAt: string; closed: boolean }>;
	reconcilingTurns: Set<string>;
};

export const emptyOverlay = (): LiveOverlay => ({
	models: new Map(),
	attempts: new Map(),
	artifacts: new Map(),
	reconcilingTurns: new Set(),
});

export const eventMatchesSession = (openSessionId: string, event: { sessionId: string }) =>
	event.sessionId === openSessionId;

// Status events and chunks use separate subscriptions and can arrive out of order.
// Either may introduce an attempt. Only a newer started_at resets its overlays;
// delayed status events must preserve current chunks, and older attempts are ignored.
// Preserve sub-millisecond precision: Date alone can merge distinct attempts.
const attemptKey = (startedAt: string) => {
	const seconds = new Date(startedAt).toISOString().slice(0, 19);
	const fraction = startedAt.match(/\.(\d+)/)?.[1] ?? "";
	return `${seconds}.${fraction.padEnd(9, "0")}Z`;
};

export const acceptAttempt = (
	overlay: LiveOverlay,
	turnId: string,
	startedAt: string
): LiveOverlay | null => {
	const key = attemptKey(startedAt);
	const previous = overlay.attempts.get(turnId);
	if (previous && key < previous.startedAt) {
		return null;
	}
	if (previous?.startedAt === key) {
		return overlay;
	}
	const attempts = new Map(overlay.attempts).set(turnId, { startedAt: key, closed: false });
	const models = new Map([...overlay.models].filter(([, model]) => model.turnId !== turnId));
	const artifacts = new Map([...overlay.artifacts].filter(([, artifact]) => artifact.turnId !== turnId));
	const reconcilingTurns = new Set(overlay.reconcilingTurns);
	reconcilingTurns.delete(turnId);
	return { models, artifacts, attempts, reconcilingTurns };
};

export const closeAttempt = (overlay: LiveOverlay, turnId: string, startedAt: string) => {
	const accepted = acceptAttempt(overlay, turnId, startedAt);
	if (!accepted) {
		return overlay;
	}
	const attempts = new Map(accepted.attempts).set(turnId, {
		startedAt: attemptKey(startedAt),
		closed: true,
	});
	return { ...accepted, attempts, reconcilingTurns: new Set(accepted.reconcilingTurns).add(turnId) };
};

export const applyModelChunk = (
	overlay: LiveOverlay,
	turnId: string,
	startedAt: string,
	chunk: AgentModelChunk
) => {
	const accepted = acceptAttempt(overlay, turnId, startedAt);
	if (!accepted || accepted.attempts.get(turnId)?.closed) {
		return overlay;
	}
	const updatedOverlay = { ...accepted, models: new Map(accepted.models) };
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

export const applyArtifactChunk = (
	overlay: LiveOverlay,
	turnId: string,
	startedAt: string,
	chunk: AgentArtifactChunk
) => {
	const accepted = acceptAttempt(overlay, turnId, startedAt);
	if (!accepted || accepted.attempts.get(turnId)?.closed) {
		return overlay;
	}
	return { ...accepted, artifacts: new Map(accepted.artifacts).set(chunk.name, { ...chunk, turnId }) };
};

export const reconcileTurn = (overlay: LiveOverlay, turnId: string, startedAt: string) => {
	if (overlay.attempts.get(turnId)?.startedAt !== attemptKey(startedAt)) {
		return overlay;
	}
	const models = new Map([...overlay.models].filter(([, model]) => model.turnId !== turnId));
	const artifacts = new Map([...overlay.artifacts].filter(([, artifact]) => artifact.turnId !== turnId));
	const reconcilingTurns = new Set(overlay.reconcilingTurns);
	reconcilingTurns.delete(turnId);
	return { ...overlay, models, artifacts, reconcilingTurns };
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
