import {
	getAiAgentRunOptions,
	type AiAgentRun,
	type AiAgentRunSnapshot,
	type ErrorModel,
	type Part,
} from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import type { BadgeVariant } from "$components/ui/badge";

const pendingStatus = "pending";

type DisplayRow = {
	label: string;
	value: string;
};

export type AgentRunJsonSection = {
	label: string;
	value: string;
};

export type AgentRunPartView = {
	text?: string;
	rows: DisplayRow[];
	jsonSections: AgentRunJsonSection[];
};

export type AgentRunMessageView = {
	role: string;
	metadata?: string;
	parts: AgentRunPartView[];
};

export type AgentRunArtifactView = {
	name: string;
	metadata?: string;
	parts: AgentRunPartView[];
};

export type AgentRunSnapshotView = {
	id: string;
	status: string;
	statusLabel: string;
	statusVariant: BadgeVariant;
	statusClass?: string;
	finishReason: string;
	parentId: string;
	heartbeatAt: string;
	createdAt: string;
	error?: string;
	stateSummary: string;
	stateCustom?: string;
	isPending: boolean;
	messages: AgentRunMessageView[];
	artifacts: AgentRunArtifactView[];
};

const statusDisplays: Record<
	string,
	Pick<AgentRunSnapshotView, "statusLabel" | "statusVariant" | "statusClass">
> = {
	pending: { statusLabel: "Pending", statusVariant: "outline", statusClass: "text-blue-700" },
	completed: { statusLabel: "Completed", statusVariant: "secondary", statusClass: "text-green-700" },
	aborted: { statusLabel: "Aborted", statusVariant: "outline" },
	failed: { statusLabel: "Failed", statusVariant: "destructive" },
};

const formatDateTime = (value?: string | null) => {
	if (!value) return "Not set";
	const date = new Date(value);
	if (Number.isNaN(date.getTime())) return value;
	return date.toLocaleString();
};

const formatJson = (value: unknown) => {
	if (value === undefined || value === null) return undefined;
	return JSON.stringify(value, null, 2);
};

const addRow = (rows: DisplayRow[], label: string, value: unknown) => {
	if (value === undefined || value === null || value === "") return;
	rows.push({ label, value: String(value) });
};

const makePartView = (part: Part): AgentRunPartView => {
	const rows: DisplayRow[] = [];
	const jsonSections: AgentRunJsonSection[] = [];

	addRow(rows, "Kind", part.kind);
	addRow(rows, "Content type", part.contentType);
	addRow(rows, "Resource URI", part.resource?.uri);

	if (part.toolRequest) {
		addRow(rows, "Tool request", part.toolRequest.name);
		addRow(rows, "Tool request ref", part.toolRequest.ref);
		addRow(rows, "Tool request partial", part.toolRequest.partial);

		const input = formatJson(part.toolRequest.input);
		if (input) jsonSections.push({ label: "Tool request input", value: input });
	}

	if (part.toolResponse) {
		addRow(rows, "Tool response", part.toolResponse.name);
		addRow(rows, "Tool response ref", part.toolResponse.ref);

		const output = formatJson(part.toolResponse.output);
		if (output) jsonSections.push({ label: "Tool response output", value: output });

		const content = formatJson(part.toolResponse.content);
		if (content) jsonSections.push({ label: "Tool response content", value: content });
	}

	const metadata = formatJson(part.metadata);
	if (metadata) jsonSections.push({ label: "Metadata", value: metadata });

	const custom = formatJson(part.custom);
	if (custom) jsonSections.push({ label: "Custom", value: custom });

	return {
		text: part.text,
		rows,
		jsonSections,
	};
};

const makeSnapshotView = (snapshot: AiAgentRunSnapshot): AgentRunSnapshotView => {
	const attrs = snapshot.attributes;
	const state = attrs.state;
	const display = statusDisplays[attrs.status] ?? {
		statusLabel: attrs.status,
		statusVariant: "outline" as const,
	};

	const messages = state?.messages ?? [];
	const artifacts = state?.artifacts ?? [];
	const stateCustom = formatJson(state?.custom);

	return {
		id: snapshot.id,
		status: attrs.status,
		statusLabel: display.statusLabel,
		statusVariant: display.statusVariant,
		statusClass: display.statusClass,
		finishReason: attrs.finish_reason || "Not set",
		parentId: attrs.parent_id || "None",
		heartbeatAt: formatDateTime(attrs.heartbeat_at),
		createdAt: formatDateTime(attrs.created_at),
		error: attrs.error,
		stateSummary: state ? `${messages.length} messages, ${artifacts.length} artifacts` : "No state",
		stateCustom,
		isPending: attrs.status === pendingStatus,
		messages: messages.map((message) => ({
			role: message.role || "unknown",
			metadata: formatJson(message.metadata),
			parts: (message.content ?? []).map(makePartView),
		})),
		artifacts: artifacts.map((artifact, index) => ({
			name: artifact.name || `Artifact ${index + 1}`,
			metadata: formatJson(artifact.metadata),
			parts: artifact.parts.map(makePartView),
		})),
	};
};

export class AiAgentRunComponentController {
	runId = $state("");
	isExpanded = $state(false);

	constructor(idFn: Getter<string>) {
		this.runId = idFn();
		watch(idFn, (id) => {
			this.runId = id;
		});
	}

	private agentRunQuery = createQuery(() => ({
		...getAiAgentRunOptions({ path: { id: this.runId } }),
		enabled: !!this.runId,
	}));

	agentRun = $derived<AiAgentRun | undefined>(this.agentRunQuery.data?.data);
	agentName = $derived(this.agentRun?.attributes.agentname ?? "AI agent");
	isLoading = $derived(this.agentRunQuery.isLoading || this.agentRunQuery.isPending);
	error = $derived(this.agentRunQuery.error as ErrorModel | undefined);

	snapshots = $derived.by<AgentRunSnapshotView[]>(() => {
		const snapshots = this.agentRun?.attributes.latestSnapshot ?? [];
		return [...snapshots]
			.sort((a, b) => Date.parse(b.attributes.created_at) - Date.parse(a.attributes.created_at))
			.map(makeSnapshotView);
	});

	latestSnapshot = $derived(this.snapshots[0]);
	isLatestPending = $derived(!!this.latestSnapshot?.isPending);
	latestStateSummary = $derived(this.latestSnapshot?.stateSummary ?? "No snapshots");

	setExpanded(open: boolean) {
		this.isExpanded = open;
	}
}

const ctx = new Context<AiAgentRunComponentController>("AiAgentRunComponentController");
export const initAiAgentRunComponentController = (idFn: Getter<string>) =>
	ctx.set(new AiAgentRunComponentController(idFn));
export const useAiAgentRunComponentController = () => ctx.get();
