import {
	createAgentCaseMutation,
	listAgentCaseArtifactsOptions,
	listAgentCasesOptions,
	listAgentRunsOptions,
	requestAgentCaseRunMutation,
	type AgentCase,
	type AgentCaseArtifact,
	type AgentRun,
} from "$lib/api";
import { WebSocketStatus } from "@hocuspocus/provider";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";
import { useIncidentCollaboration } from "../collaboration.svelte";
import { useIncidentView } from "../controller.svelte";

type Drawer = "add-component";
type Payload = Record<string, unknown> | undefined;

type LinkedItem = {
	id: string;
	title: string;
	summary: string;
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

	private casesQueryOptions = $derived(
		listAgentCasesOptions({
			query: {
				limit: 5,
				workflowKind: "incident_context_pack",
				subjectKind: "incident",
				subjectId: this.incidentId,
			},
		})
	);
	casesQuery = createQuery(() => ({
		...this.casesQueryOptions,
		enabled: !!this.incidentId,
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
	explicitImpacts = $derived(this.artifactsByRole("explicit_impact"));
	inferredImpacts = $derived(this.artifactsByRole("inferred_impact"));
	activeAlerts = $derived(this.artifactsByRole("active_alert"));
	relatedIncidents = $derived(this.artifactsByRole("related_incident"));
	impacts = $derived([...this.explicitImpacts, ...this.inferredImpacts]);
	latestRunStatusLabel = $derived(this.latestRun?.attributes.status.replaceAll("_", " "));
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
		},
	}));

	requestPending = $derived(this.createCase.isPending || this.requestRun.isPending);
	requestDisabled = $derived(!this.incidentId || this.requestPending || this.latestRunActive);

	requestContextPack = () => {
		if (!this.incidentId) return;
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
					title: "Incident context pack",
					query: "Build current triage context for the incident.",
					workflowKind: "incident_context_pack",
					subjectKind: "incident",
					subjectId: this.incidentId,
					triggerMetadata: { trigger: "manual" },
				},
			},
		});
	};

	private scorePercent(score: number | undefined) {
		return `${Math.round((score ?? 0) * 100)}%`;
	}

	private stringField(payload: Payload, key: string) {
		const value = payload?.[key];
		return typeof value === "string" ? value : "";
	}

	private numberField(payload: Payload, key: string) {
		const value = payload?.[key];
		return typeof value === "number" ? value : undefined;
	}

	private artifactString(artifact: AgentCaseArtifact, key: string) {
		return this.stringField(artifact.attributes.payload, key);
	}

	private artifactsByRole(role: string) {
		return this.artifacts.filter((artifact) => artifact.attributes.role === role);
	}

	private linkedItem(artifact: AgentCaseArtifact): LinkedItem {
		return {
			id: artifact.id,
			title: this.artifactString(artifact, "title") || artifact.attributes.name,
			summary: this.artifactString(artifact, "summary"),
		};
	}
}

const ctx = new Context<IncidentContextSidebarController>("IncidentContextSidebarController");
export const initIncidentContextSidebarController = () => ctx.set(new IncidentContextSidebarController());
export const useIncidentContextSidebarController = () => ctx.get();
