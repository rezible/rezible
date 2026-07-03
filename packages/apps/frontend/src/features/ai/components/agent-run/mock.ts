import type { AiAgentRun } from "$lib/api";

const now = new Date("2026-07-03T04:00:00.000Z");
const isoMinutesAgo = (minutes: number) => new Date(now.getTime() - minutes * 60_000).toISOString();

export const mockAiAgentRun: AiAgentRun = {
	id: "9b8d2749-59cc-4e0d-9b78-98ae4e1cf501",
	attributes: {
		agentname: "Incident Investigation Agent",
		ownerUserId: "8c8c1b3d-6d54-4f31-8875-c8913f7e4ad7",
		permissionScopes: ["incidents:read", "events:read", "topology:read"],
		createdAt: isoMinutesAgo(12),
		startedAt: isoMinutesAgo(11),
		latestSnapshot: [
			{
				id: "5ecdf9c2-7e81-49df-8350-80594169de89",
				attributes: {
					status: "pending",
					finish_reason: "",
					parent_id: "71fe66df-9767-4d77-b945-88b612d54d80",
					heartbeat_at: isoMinutesAgo(1),
					created_at: isoMinutesAgo(1),
					state: {
						custom: {
							phase: "correlating-events",
							progress: 0.72,
							currentStep: "Comparing deploy and alert timelines",
						},
						messages: [
							{
								role: "user",
								content: [
									{
										text: "Investigate the checkout latency spike and summarize likely causes.",
										contentType: "text/plain",
									},
								],
								metadata: {
									source: "incident-sidebar",
								},
							},
							{
								role: "model",
								content: [
									{
										text: "I found a deploy shortly before the latency increase and am checking related service events.",
										contentType: "text/plain",
									},
									{
										kind: 3,
										toolRequest: {
											name: "search_events",
											ref: "toolu_01",
											partial: false,
											input: {
												query: "checkout latency deploy errors",
												window: "2h",
											},
										},
									},
								],
							},
						],
						artifacts: [
							{
								name: "Timeline evidence",
								metadata: {
									sourceCount: 3,
									confidence: "medium",
								},
								parts: [
									{
										text: "10:42 deploy completed, 10:47 checkout p95 latency breached alert threshold.",
										contentType: "text/plain",
									},
									{
										resource: {
											uri: "rezible://events/checkout-latency-window",
										},
										metadata: {
											label: "Checkout latency event window",
										},
									},
								],
							},
						],
					},
				},
			},
			{
				id: "71fe66df-9767-4d77-b945-88b612d54d80",
				attributes: {
					status: "completed",
					finish_reason: "stop",
					parent_id: "",
					heartbeat_at: isoMinutesAgo(6),
					created_at: isoMinutesAgo(6),
					state: {
						custom: {
							phase: "initial-context",
							progress: 0.35,
						},
						messages: [
							{
								role: "model",
								content: [
									{
										text: "I loaded incident context, related alerts, and recent topology changes.",
										contentType: "text/plain",
									},
									{
										kind: 4,
										toolResponse: {
											name: "load_incident_context",
											ref: "toolu_00",
											output: {
												alerts: 2,
												events: 18,
												relatedServices: ["checkout-api", "payments-api"],
											},
											content: [
												{
													text: "Context loaded successfully.",
													contentType: "text/plain",
												},
											],
										},
									},
								],
							},
						],
						artifacts: [
							{
								name: "Context summary",
								parts: [
									{
										text: "Primary service: checkout-api. Related signals: latency, deploy, error-rate.",
										contentType: "text/plain",
									},
								],
							},
						],
					},
				},
			},
		],
	},
};

export const mockAiAgentRunId = mockAiAgentRun.id;
