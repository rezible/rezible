import type { AgentSession, AgentTurn } from "$lib/api";

const now = new Date("2026-07-03T04:00:00.000Z");
const isoMinutesAgo = (minutes: number) => new Date(now.getTime() - minutes * 60_000).toISOString();

export const mockAgentTurns: AgentTurn[] = [
	{
		id: "5ecdf9c2-7e81-49df-8350-80594169de89",
		attributes: {
			status: "completed",
			finishReason: "stop",
			parentTurnId: "71fe66df-9767-4d77-b945-88b612d54d80",
			createdAt: isoMinutesAgo(2),
			updatedAt: isoMinutesAgo(1),
			state: {
				sessionId: "mock",
				messages: [{ role: "model", content: [{ text: "Investigation complete." }] }],
				custom: { phase: "complete" },
			},
		},
	},
	{
		id: "71fe66df-9767-4d77-b945-88b612d54d80",
		attributes: {
			status: "completed",
			finishReason: "stop",
			createdAt: isoMinutesAgo(12),
			updatedAt: isoMinutesAgo(6),
			state: { sessionId: "mock", messages: [], custom: { phase: "initial-context" } },
		},
	},
];

export const mockAgentSession: AgentSession = {
	id: "mock",
	attributes: {
		agentName: "Incident Investigation Agent",
		ownerUserId: "8c8c1b3d-6d54-4f31-8875-c8913f7e4ad7",
		permissionScopes: ["incidents:read", "events:read", "topology:read"],
		createdAt: isoMinutesAgo(12),
		latestTurn: mockAgentTurns[0],
	},
};
