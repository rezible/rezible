import {
	createAgentTaskMutation,
	listAgentRunToolCallsOptions,
	listAgentRunsOptions,
	listAgentTasksOptions,
	requestAgentTaskRunMutation,
	type AgentRun,
	type AgentRunToolCall,
	type AgentTask,
} from "$lib/api";
import { WebSocketStatus } from "@hocuspocus/provider";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";
import { useIncidentCollaboration } from "../collaboration.svelte";
import { useIncidentView } from "../controller.svelte";
import { getAgentRunStatus, isAgentRunActive } from "$src/lib/agents.svelte";

type Drawer = "add-component";
type Payload = Record<string, unknown> | undefined;

type LinkedItem = {
	id: string;
	title: string;
	summary: string;
};

type ContextItem = {
	id: string;
	attributes: {
		role: string;
		name: string;
		payload?: Record<string, unknown>;
	};
};

export class IncidentContextSidebarController {
	private view = useIncidentView();
	private collab = useIncidentCollaboration();
	private queryClient = useQueryClient();

	drawer = $state<Drawer>();
	drawerOpen = $derived(!!this.drawer);

	ctxColor = $derived.by(() => {
		if (this.collab.error) return "fill-danger";
		switch (this.collab.status) {
			case WebSocketStatus.Connecting:
				return "fill-default";
			case WebSocketStatus.Connected:
				return "fill-success";
			case WebSocketStatus.Disconnected:
				return "fill-warning";
		}
	});
	connectionError = $derived(this.collab.error);

	incidentId = $derived(this.view.incidentId);

	private tasksQueryOptions = $derived(
		listAgentTasksOptions({
			query: {
				limit: 5,
				workflow: "incident_context_pack",
				subjectKind: "incident",
				domainEntityId: this.incidentId,
			},
		})
	);
	tasksQuery = createQuery(() => ({
		...this.tasksQueryOptions,
		enabled: !!this.incidentId,
		refetchInterval: 3000,
	}));
	casesQuery = this.tasksQuery;
	latestTask: AgentTask | undefined = $derived(this.tasksQuery.data?.data?.[0]);
	latestTaskId = $derived(this.latestTask?.id ?? "");

	private runsQueryOptions = $derived(
		listAgentRunsOptions({
			query: {
				limit: 5,
				agentTaskId: this.latestTaskId,
			},
		})
	);
	private runsQuery = createQuery(() => ({
		...this.runsQueryOptions,
		enabled: !!this.latestTaskId,
		refetchInterval: 3000,
	}));
	latestRun: AgentRun | undefined = $derived(this.runsQuery.data?.data?.[0]);
	private latestRunActive = $derived(!!this.latestRun && isAgentRunActive(this.latestRun));

	private toolCallsQueryOptions = $derived(
		listAgentRunToolCallsOptions({ path: { id: this.latestRun?.id ?? "" } })
	);
	private toolCallsQuery = createQuery(() => ({
		...this.toolCallsQueryOptions,
		enabled: !!this.latestRun?.id,
		refetchInterval: this.latestRunActive ? 3000 : false,
	}));
	private toolCalls: AgentRunToolCall[] = $derived(this.toolCallsQuery.data?.data ?? []);
	private contextToolCall: AgentRunToolCall | undefined = $derived(
		this.toolCalls.find((call) => call.attributes.toolId === "incident.context_pack")
	);
	artifacts: ContextItem[] = $derived(this.contextItems(this.contextToolCall));
	explicitImpacts = $derived(this.artifactsByRole("explicit_impact"));
	inferredImpacts = $derived(this.artifactsByRole("inferred_impact"));
	activeAlerts = $derived(this.artifactsByRole("active_alert"));
	relatedIncidents = $derived(this.artifactsByRole("related_incident"));
	impacts = $derived([...this.explicitImpacts, ...this.inferredImpacts]);
	latestRunStatusLabel = $derived(getAgentRunStatus(this.latestRun));
	impactItems = $derived(
		this.impacts.slice(0, 5).map((impact) => ({
			id: impact.id,
			displayName: this.artifactString(impact, "displayName") || impact.attributes.name,
			kind: this.artifactString(impact, "kind"),
			score: this.scorePercent(this.numberField(impact.attributes.payload, "score")),
			reason: this.artifactString(impact, "reason"),
		}))
	);
	activeAlertItems = $derived(
		this.activeAlerts.slice(0, 4).map((alert) => ({
			...this.linkedItem(alert),
			routeId: this.artifactString(alert, "id"),
		}))
	);
	relatedIncidentItems = $derived(
		this.relatedIncidents.slice(0, 4).map((incident) => ({
			...this.linkedItem(incident),
			routeSlug: this.artifactString(incident, "source"),
		}))
	);

	private createTask = createMutation(() => ({
		...createAgentTaskMutation(),
		onSuccess: async () => {
			await this.queryClient.invalidateQueries(this.tasksQueryOptions);
			await this.queryClient.invalidateQueries(this.runsQueryOptions);
		},
	}));
	private requestRun = createMutation(() => ({
		...requestAgentTaskRunMutation(),
		onSuccess: async () => {
			await this.queryClient.invalidateQueries(this.runsQueryOptions);
			await this.queryClient.invalidateQueries(this.toolCallsQueryOptions);
		},
	}));

	requestPending = $derived(this.createTask.isPending || this.requestRun.isPending);
	requestDisabled = $derived(!this.incidentId || this.requestPending || this.latestRunActive);

	requestContextPack = () => {
		if (!this.incidentId) return;
		if (this.latestTaskId) {
			this.requestRun.mutate({
				path: { id: this.latestTaskId },
			});
			return;
		}

		this.createTask.mutate({
			body: {
				attributes: {
					workflow: "incident_context_pack",
					input: {
						schema: "incident_context_pack.v1",
						subjects: [{ type: "incident", id: this.incidentId }],
						objectives: ["Build current triage context for the incident."],
					},
					triggerKind: "manual",
					triggerPayload: { trigger: "manual" },
				},
			},
		});
	};

	private scorePercent(score: number | undefined) {
		const value = score ?? 0;
		return `${Math.round(value > 1 ? value : value * 100)}%`;
	}

	private stringField(payload: Payload, key: string) {
		const value = payload?.[key];
		return typeof value === "string" ? value : "";
	}

	private numberField(payload: Payload, key: string) {
		const value = payload?.[key];
		return typeof value === "number" ? value : undefined;
	}

	private artifactString(artifact: ContextItem, key: string) {
		return this.stringField(artifact.attributes.payload, key);
	}

	private artifactsByRole(role: string) {
		return this.artifacts.filter((artifact) => artifact.attributes.role === role);
	}

	private linkedItem(artifact: ContextItem): LinkedItem {
		return {
			id: artifact.id,
			title: this.artifactString(artifact, "title") || artifact.attributes.name,
			summary: this.artifactString(artifact, "summary"),
		};
	}

	private contextItems(toolCall: AgentRunToolCall | undefined): ContextItem[] {
		const rawItems = toolCall?.attributes.result?.items;
		if (!Array.isArray(rawItems)) return [];
		return rawItems
			.filter((item): item is Record<string, unknown> => item !== null && typeof item === "object")
			.map((item, index) => {
				const payload = this.recordField(item as Payload, "payload");
				return {
					id: this.stringField(item as Payload, "id") || `${index}`,
					attributes: {
						role: this.stringField(item as Payload, "role"),
						name: this.stringField(item as Payload, "name"),
						payload,
					},
				};
			});
	}

	private recordField(payload: Payload, key: string) {
		const value = payload?.[key];
		return value && typeof value === "object" && !Array.isArray(value)
			? (value as Record<string, unknown>)
			: undefined;
	}
}

const ctx = new Context<IncidentContextSidebarController>("IncidentContextSidebarController");
export const initIncidentContextSidebarController = () => ctx.set(new IncidentContextSidebarController());
export const useIncidentContextSidebarController = () => ctx.get();
