import {
	createAgentCaseMutation,
	listAgentCaseArtifactsOptions,
	listAgentCaseConclusionsOptions,
	listAgentCasesOptions,
	listAgentRunsOptions,
	requestAgentCaseRunMutation,
	type AgentCase,
	type AgentCaseArtifact,
	type AgentCaseConclusion,
	type AgentRun,
} from "$lib/api";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";
import { useAlertViewController } from "../controller.svelte";

type Payload = Record<string, unknown> | undefined;

type TitledItem = {
	id: string;
	title: string;
	summary: string;
};

export class AlertInvestigationController {
	private view = useAlertViewController();
	private queryClient = useQueryClient();

	alertId = $derived(this.view.alertId);

	private casesQueryOptions = $derived(
		listAgentCasesOptions({
			query: {
				limit: 5,
				workflowKind: "alert_investigation",
				subjectKind: "alert",
				subjectId: this.alertId,
			},
		})
	);
	casesQuery = createQuery(() => ({
		...this.casesQueryOptions,
		enabled: !!this.alertId,
		refetchInterval: 3000,
	}));
	latestCase: AgentCase | undefined = $derived(this.casesQuery.data?.data?.[0]);
	latestCaseId = $derived(this.latestCase?.id ?? "");

	private runsQueryOptions = $derived(
		listAgentRunsOptions({
			query: {
				limit: 5,
				agentCaseId: this.latestCaseId,
			},
		})
	);
	private runsQuery = createQuery(() => ({
		...this.runsQueryOptions,
		enabled: !!this.latestCaseId,
		refetchInterval: 3000,
	}));
	latestRun: AgentRun | undefined = $derived(this.runsQuery.data?.data?.[0]);
	private latestRunActive = $derived(
		this.latestRun?.attributes.status === "queued" || this.latestRun?.attributes.status === "running"
	);

	private artifactsQueryOptions = $derived(
		listAgentCaseArtifactsOptions({ path: { id: this.latestCaseId } })
	);
	private artifactsQuery = createQuery(() => ({
		...this.artifactsQueryOptions,
		enabled: !!this.latestCaseId,
		refetchInterval: this.latestRunActive ? 3000 : false,
	}));
	artifacts = $derived(this.artifactsQuery.data?.data ?? []);

	private conclusionsQueryOptions = $derived(
		listAgentCaseConclusionsOptions({ path: { id: this.latestCaseId } })
	);
	private conclusionsQuery = createQuery(() => ({
		...this.conclusionsQueryOptions,
		enabled: !!this.latestCaseId,
		refetchInterval: this.latestRunActive ? 3000 : false,
	}));

	contextArtifact: AgentCaseArtifact | undefined = $derived(
		this.artifacts.find(
			(artifact) =>
				artifact.attributes.kind === "context" && artifact.attributes.name === "retrieval_context"
		)
	);
	context = $derived(this.contextArtifact?.attributes.payload);
	resultConclusion: AgentCaseConclusion | undefined = $derived(
		this.conclusionsQuery.data?.data.find(
			(conclusion) => conclusion.attributes.kind === "alert_investigation"
		)
	);
	resultPayload = $derived(this.resultConclusion?.attributes.payload);
	findingsPayload = $derived(this.recordField(this.resultPayload, "findings"));

	private createCase = createMutation(() => ({
		...createAgentCaseMutation(),
		onSuccess: async () => {
			await this.queryClient.invalidateQueries(this.casesQueryOptions);
			await this.queryClient.invalidateQueries(this.runsQueryOptions);
		},
	}));
	private requestRun = createMutation(() => ({
		...requestAgentCaseRunMutation(),
		onSuccess: async () => {
			await this.queryClient.invalidateQueries(this.runsQueryOptions);
			await this.queryClient.invalidateQueries(this.artifactsQueryOptions);
			await this.queryClient.invalidateQueries(this.conclusionsQueryOptions);
		},
	}));

	requestPending = $derived(this.createCase.isPending || this.requestRun.isPending);
	requestDisabled = $derived(this.requestPending || this.latestRunActive);

	subjects = $derived(this.artifactsByRole("likely_subject"));
	signals = $derived(this.artifactsByRole("recent_signal"));
	suggestedChecks = $derived(this.stringList(this.context, "suggestedChecks"));
	neighbors = $derived(this.artifactsByRole("neighbor"));
	guides = $derived(this.artifactsByRole("guide"));
	title = $derived(
		this.stringField(this.context, "alertTitle") ||
			this.latestCase?.attributes.title ||
			this.statusLabel(this.latestRun)
	);
	resultSummary = $derived(
		this.stringField(this.resultPayload, "summary") || this.resultConclusion?.attributes.summary || ""
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
		if (this.latestCaseId) {
			this.requestRun.mutate({
				path: { id: this.latestCaseId },
				body: { attributes: { metadata: { trigger: "manual" } } },
			});
			return;
		}

		this.createCase.mutate({
			body: {
				attributes: {
					title: "Alert investigation",
					query: "Investigate alert and summarize findings for responders.",
					workflowKind: "alert_investigation",
					subjectKind: "alert",
					subjectId: this.alertId,
					triggerMetadata: { trigger: "manual" },
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

	private artifactString(artifact: AgentCaseArtifact, key: string) {
		return this.stringField(artifact.attributes.payload, key);
	}

	statusLabel(run: AgentRun | undefined) {
		if (!run) return "No run yet";
		return run.attributes.status.replaceAll("_", " ");
	}

	private artifactsByRole(role: string) {
		return this.artifacts.filter((artifact) => artifact.attributes.role === role);
	}

	private titledItem(artifact: AgentCaseArtifact): TitledItem {
		return {
			id: artifact.id,
			title: this.artifactString(artifact, "title") || artifact.attributes.name,
			summary: this.artifactString(artifact, "summary"),
		};
	}
}

const ctx = new Context<AlertInvestigationController>("AlertInvestigationController");
export const initAlertInvestigationController = () => ctx.set(new AlertInvestigationController());
export const useAlertInvestigationController = () => ctx.get();
