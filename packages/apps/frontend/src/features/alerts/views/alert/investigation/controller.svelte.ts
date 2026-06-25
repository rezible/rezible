import {
	createAgentTaskMutation,
	getAgentRunResultOptions,
	listAgentRunToolCallsOptions,
	listAgentRunsOptions,
	listAgentTasksOptions,
	requestAgentTaskRunMutation,
	type AgentRun,
	type AgentRunResult,
	type AgentRunToolCall,
	type AgentTask,
} from "$lib/api";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";
import { useAlertViewController } from "../controller.svelte";
import { getAgentRunStatus, isAgentRunActive } from "$src/lib/agents.svelte";

type Payload = Record<string, unknown> | undefined;

type TitledItem = {
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

export class AlertInvestigationController {
	private view = useAlertViewController();
	private queryClient = useQueryClient();

	alertId = $derived(this.view.alertId);

	private tasksQueryOptions = $derived(
		listAgentTasksOptions({
			query: {
				limit: 5,
				workflow: "alert_investigation",
				subjectKind: "alert",
				domainEntityId: this.alertId,
			},
		})
	);
	tasksQuery = createQuery(() => ({
		...this.tasksQueryOptions,
		enabled: !!this.alertId,
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
		this.toolCalls.find((call) => call.attributes.toolId === "alert.investigation_context")
	);
	context = $derived(this.recordField(this.contextToolCall?.attributes.result, "context"));
	artifacts: ContextItem[] = $derived(this.contextItems(this.contextToolCall));
	result = $derived(this.latestRun?.attributes.result);
	resultPayload = $derived(this.result?.attributes.output);
	findingsPayload = $derived(this.recordField(this.resultPayload, "findings"));

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
	requestDisabled = $derived(this.requestPending || this.latestRunActive);

	subjects = $derived(this.artifactsByRole("likely_subject"));
	signals = $derived(this.artifactsByRole("recent_signal"));
	suggestedChecks = $derived(this.stringList(this.context, "suggestedChecks"));
	neighbors = $derived(this.artifactsByRole("neighbor"));
	guides = $derived(this.artifactsByRole("guide"));
	title = $derived(this.stringField(this.context, "alertTitle") || "Alert investigation");
	resultSummary = $derived(
		this.stringField(this.resultPayload, "summary") || ""
	);
	likelyCause = $derived(this.stringField(this.findingsPayload, "likelyCause"));
	findingSuggestedChecks = $derived(this.stringList(this.findingsPayload, "suggestedChecks"));
	recommendedNext = $derived(this.stringField(this.findingsPayload, "recommendedNext"));
	subjectItems = $derived(
		this.subjects.map((subject) => ({
			id: subject.id,
			displayName: this.artifactString(subject, "displayName") || subject.attributes.name,
			kind: this.artifactString(subject, "kind"),
			confidence: this.artifactString(subject, "confidence"),
			reason: this.artifactString(subject, "reason"),
		}))
	);
	signalItems = $derived(
		this.signals.slice(0, 6).map((signal) => ({
			id: signal.id,
			summary: this.artifactString(signal, "summary") || signal.attributes.name,
			source: this.artifactString(signal, "source"),
			kind: this.artifactString(signal, "kind"),
		}))
	);
	neighborItems = $derived(
		this.neighbors.slice(0, 8).map((neighbor) => ({
			id: neighbor.id,
			name: this.artifactString(neighbor, "relatedEntity") || neighbor.attributes.name,
			kind: this.artifactString(neighbor, "kind"),
			relatedEntityKind: this.artifactString(neighbor, "relatedEntityKind"),
		}))
	);
	guideItems = $derived(this.guides.map((guide): TitledItem => this.titledItem(guide)));

	runInvestigation = () => {
		if (this.latestTaskId) {
			this.requestRun.mutate({
				path: { id: this.latestTaskId },
			});
			return;
		}

		this.createTask.mutate({
			body: {
				attributes: {
					workflow: "alert_investigation",
					input: {
						schema: "alert_investigation.v1",
						subjects: [{ type: "alert", id: this.alertId }],
						objectives: ["Investigate alert and summarize findings for responders."],
					},
					triggerKind: "manual",
					triggerPayload: { trigger: "manual" },
				},
			},
		});
	};

	private recordField(payload: Payload, key: string) {
		const value = payload?.[key];
		return value && typeof value === "object" && !Array.isArray(value)
			? (value as Record<string, unknown>)
			: undefined;
	}

	private stringField(payload: Payload, key: string) {
		const value = payload?.[key];
		return typeof value === "string" ? value : "";
	}

	private stringList(payload: Payload, key: string) {
		const value = payload?.[key];
		return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string") : [];
	}

	private artifactString(artifact: ContextItem, key: string) {
		return this.stringField(artifact.attributes.payload, key);
	}

	statusLabel(run: AgentRun | undefined) {
		if (!run) return "No run yet";
		return getAgentRunStatus(run);
	}

	private artifactsByRole(role: string) {
		return this.artifacts.filter((artifact) => artifact.attributes.role === role);
	}

	private titledItem(artifact: ContextItem): TitledItem {
		return {
			id: artifact.id,
			title: this.artifactString(artifact, "title") || artifact.attributes.name,
			summary: this.artifactString(artifact, "summary"),
		};
	}

	private contextItems(toolCall: AgentRunToolCall | undefined): ContextItem[] {
		const result = toolCall?.attributes.result;
		const rawItems = result?.items;
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
}

const ctx = new Context<AlertInvestigationController>("AlertInvestigationController");
export const initAlertInvestigationController = () => ctx.set(new AlertInvestigationController());
export const useAlertInvestigationController = () => ctx.get();
