<script lang="ts">
	import { Button } from "$components/ui/button";
	import type {
		ExpandableIncidentSeverityAttributes,
		IncidentAttributes,
		InvestigationAttributes,
		RetrospectiveAttributes,
		SituationAttributes,
	} from "$lib/api";
	import type { StatusPresentation } from "$components/common/status-badge/status";
	import { formatDuration } from "$lib/time";
	import * as Card from "$components/ui/card";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import {
		hypothesisStatus,
		investigationRunStatus,
		situationStatus,
	} from "$features/situations/lib/status";
	import {
		incidentResponseStatus,
		incidentSeverityStatus,
		retrospectiveStatus,
	} from "$features/incidents/lib/status";

	// Illustrative inputs only; these are not product data.
	const situation = (attributes: Partial<SituationAttributes>) => attributes as SituationAttributes;
	const investigation = (attributes: Partial<InvestigationAttributes>) =>
		attributes as InvestigationAttributes;
	const turn = (status: "queued" | "running" | "completed" | "failed" | "aborted") => ({
		id: "illustrative-turn",
		status,
	});
	const severity = (name: string, rank: number): ExpandableIncidentSeverityAttributes => ({
		id: name,
		attributes: { name, rank, archived: false, color: "", description: "" },
	});

	const responseStates: IncidentAttributes["responseState"][] = [
		"started",
		"mitigated",
		"resolved",
		"unknown",
	];
	const reviewStates: (RetrospectiveAttributes["state"] | undefined)[] = [
		undefined,
		"draft",
		"in_review",
		"meeting",
		"closed",
	];

	const groups: { label: string; statuses: StatusPresentation[] }[] = [
		{
			label: "Situation",
			statuses: [
				situationStatus(situation({})),
				situationStatus(situation({ investigation: { investigation: { id: "illustrative" } } })),
				situationStatus(situation({ closedAt: "2026-05-14T04:45:00Z", closeReason: "stabilized" })),
			],
		},
		{
			label: "Investigation run",
			statuses: [
				investigationRunStatus(undefined),
				investigationRunStatus(investigation({ activeTurn: turn("queued") })),
				investigationRunStatus(investigation({ activeTurn: turn("running") })),
				investigationRunStatus(investigation({ latestTurn: turn("completed") })),
				investigationRunStatus(investigation({ latestTurn: turn("failed") })),
				investigationRunStatus(investigation({ latestTurn: turn("aborted") })),
			],
		},
		{
			label: "Hypothesis",
			statuses: ["open", "supported", "disproven", "inconclusive"].map(hypothesisStatus),
		},
		{ label: "Incident response", statuses: responseStates.map(incidentResponseStatus) },
		{ label: "Review", statuses: reviewStates.map(retrospectiveStatus) },
		{
			label: "Severity",
			statuses: [
				incidentSeverityStatus(severity("SEV-1", 1)),
				incidentSeverityStatus(severity("SEV-2", 2)),
				incidentSeverityStatus(severity("SEV-3", 3)),
				incidentSeverityStatus(undefined),
			],
		},
	];

	const now = Date.now();
	const ago = (ms: number) => new Date(now - ms).toISOString();
	const times = [
		{ label: "30 seconds ago", value: ago(30_000) },
		{ label: "5 minutes ago", value: ago(5 * 60_000) },
		{ label: "3 hours ago", value: ago(3 * 3_600_000) },
		{ label: "4 days ago", value: ago(4 * 86_400_000) },
		{ label: "40 days ago", value: ago(40 * 86_400_000) },
		{ label: "Invalid", value: "0001-01-01T00:00:00Z" },
	];

	const durationStart = "2026-05-14T00:00:00Z";
	const after = (ms: number) => new Date(Date.parse(durationStart) + ms).toISOString();
	const durations = [
		formatDuration(durationStart, after(45_000)),
		formatDuration(durationStart, after(23 * 60_000)),
		formatDuration(durationStart, after(4 * 3_600_000 + 5 * 60_000)),
		formatDuration(durationStart, after(3 * 86_400_000 + 2 * 3_600_000)),
	];
</script>

{#snippet viewAll()}
	<Button variant="link" size="inline" href="#status-and-time">View all</Button>
{/snippet}

<Card.Root id="status-and-time">
	<Card.Header>
		<Card.Title>Status and time</Card.Title>
		<Card.Description>
			Shared status presenters, timestamps and section headings. Inputs are illustrative.
		</Card.Description>
	</Card.Header>
	<Card.Content class="flex flex-col gap-6">
		{#each groups as group (group.label)}
			<div class="flex flex-col gap-2">
				<p class="region-label">
					{group.label}
				</p>
				<div class="flex flex-wrap items-center gap-2">
					{#each group.statuses as status (status.label)}
						<StatusBadge {status} />
					{/each}
				</div>
				<div class="flex flex-wrap items-center gap-4">
					{#each group.statuses as status (status.label)}
						<StatusBadge {status} variant="inline" />
					{/each}
				</div>
			</div>
		{/each}

		<div class="flex flex-col gap-2">
			<p class="region-label">Timestamps</p>
			<table class="text-sm">
				<thead class="text-left text-xs text-muted-foreground">
					<tr>
						<th class="py-1 pe-4 font-medium">Input</th>
						<th class="py-1 pe-4 font-medium">Relative</th>
						<th class="py-1 pe-4 font-medium">Clock</th>
						<th class="py-1 font-medium">Absolute</th>
					</tr>
				</thead>
				<tbody>
					{#each times as time (time.label)}
						<tr>
							<td class="py-1 pe-4 text-muted-foreground">{time.label}</td>
							<td class="py-1 pe-4"><Timestamp value={time.value} /></td>
							<td class="py-1 pe-4"><Timestamp value={time.value} format="clock" /></td>
							<td class="py-1"><Timestamp value={time.value} format="absolute" /></td>
						</tr>
					{/each}
				</tbody>
			</table>
			<p class="text-sm tabular-nums">Durations: {durations.join(" · ")}</p>
		</div>

		<div class="flex flex-col gap-4">
			<SectionHeading
				id="specimen-section-heading"
				title="Section heading"
				count={12}
				actions={viewAll}
			/>
			<SectionHeading id="specimen-subsection-heading" title="Subsection heading" count={3} level={3} />
		</div>
	</Card.Content>
</Card.Root>
