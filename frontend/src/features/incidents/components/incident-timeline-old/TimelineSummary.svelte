<script lang="ts">
	import { cls } from '@layerstack/tailwind';
	import { Icon, Tooltip } from "svelte-ux";
	import type { IncidentEvent, IncidentMilestone, IncidentStage } from "./events";
	import {
		mdiAccountAlert,
		mdiAlarmLight,
		mdiChartLine,
		mdiFire,
		mdiFlag,
		mdiMagnify,
		mdiProgressWrench,
		mdiThoughtBubble,
		mdiWrench,
	} from "@mdi/js";
	import { differenceInSeconds } from "date-fns";

	type Props = {
		events: IncidentEvent[];
		selectedId?: string;
		hoveringId?: string;
		onEventClicked?: (eventId: string) => void;
	};
	let { events, selectedId, hoveringId, onEventClicked = (id) => {} }: Props = $props();

	type StageMarker = { stage: IncidentStage; start: Date; chunk: number };
	const getStageMarkers = (events: IncidentEvent[]) => {
		let stageEvents: IncidentEvent[] = [];
		for (let i = 0; i < events.length; i++) {
			const e = events[i];
			if (e.stage_change === undefined) continue;
			stageEvents.push(e); //{stage: e.stage_change, time: e.start});
		}

		let markers: StageMarker[] = [];

		const start = events[0].start;
		const lastEvent = events[events.length - 1];
		const end = lastEvent.end ?? lastEvent.start;
		const timelineDuration = differenceInSeconds(end, start);

		let chunksTotal = 0;
		for (let i = 0; i < stageEvents.length; i++) {
			const e = stageEvents[i];

			let chunk = 0;
			if (i + 1 === stageEvents.length) {
				chunk = 1 - chunksTotal;
			} else {
				chunk = differenceInSeconds(stageEvents[i + 1].start, e.start) / timelineDuration;
			}
			chunksTotal += chunk;

			if (e.stage_change) {
				markers.push({
					stage: e.stage_change,
					start: e.start,
					chunk: Math.round(100 * chunk),
				});
			}
		}
		return markers;
	};

	const icons: Record<IncidentMilestone, string> = {
		impact_start: mdiFire,
		metrics: mdiChartLine,
		alert: mdiAlarmLight,
		response_start: mdiAccountAlert,
		incident_detail: mdiMagnify,
		hypothesis: mdiThoughtBubble,
		mitigation_attempt: mdiProgressWrench,
		mitigated: mdiWrench,
	};

	type FormatEvent = {
		id: string;
		title: string;
		icon: string;
		position: number;
	};
	const getFormattedEvents = (events: IncidentEvent[]): FormatEvent[] => {
		let formatted: FormatEvent[] = [];

		const start = events[0].start;
		const lastEvent = events[events.length - 1];
		const end = lastEvent.end ?? lastEvent.start;
		const timelineDuration = differenceInSeconds(end, start);

		for (let i = 0; i < events.length; i++) {
			const event = events[i];
			const position = Math.round((differenceInSeconds(event.start, start) / timelineDuration) * 100);

			let title = "";
			let icon = mdiFlag;
			if (event.type === "milestone") {
				title = event.milestone;
				icon = icons[event.milestone];
			} else {
				title = event.description;
			}
			formatted.push({ id: event.id, title, icon, position });
		}

		return formatted;
	};

	const stageColors: Record<IncidentStage, string> = {
		impact: "border-danger-900/50 bg-danger-900/20",
		detection: "border-warning-600/50 bg-warning-600/20",
		response: "border-info-700/50 bg-info-700/20",
		mitigation: "border-success-600/50 bg-success-600/20",
	};
</script>

<div class="w-full min-h-18 h-fit flex flex-col align-center pb-1">
	<div class="w-full flex">
		<div class="hidden bg-error-600/50 bg-warning-600/50 bg-info-700/50 bg-success-600/50"></div>
		{#each getStageMarkers(events) as s}
			<span
				class="inline-block px-1 leading-2 border-s-4 {stageColors[s.stage]}"
				style="width: {s.chunk}%"
			>
				{s.stage}
			</span>
		{/each}
	</div>
	<div class="w-full my-1 h-[32px] relative flex items-center">
		<div class="mx-1 w-full h-2 bg-neutral-300/50"></div>

		{#each getFormattedEvents(events) as e (e.id)}
			<div
				class={cls(
					"rounded-full border absolute",
					e.id === hoveringId ? "bg-secondary-900" : "bg-surface-300",
					e.id === selectedId ? "bg-secondary-800" : ""
				)}
				style="left: clamp(0%, {e.position}%, calc(100% - 32px))"
			>
				<Tooltip title={e.title}>
					<button class="size-8" onclick={() => onEventClicked(e.id)}>
						<Icon size={16} data={e.icon} />
					</button>
				</Tooltip>
			</div>
		{/each}
	</div>
</div>
