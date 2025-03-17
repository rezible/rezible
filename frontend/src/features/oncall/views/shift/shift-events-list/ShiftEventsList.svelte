<script lang="ts">
	import { mdiPhoneAlert, mdiFire, mdiCheckCircle, mdiClockOutline, mdiAlertCircle } from "@mdi/js";
	import { Icon, Header, Badge } from "svelte-ux";
	import { settings } from "$lib/settings.svelte";
	import type { ShiftEvent } from "$features/oncall/lib/utils";
	import { PeriodType } from "@layerstack/utils";
	import { formatDistanceToNow } from "date-fns";
	import type { ZonedDateTime } from "@internationalized/date";

	type Props = {
		shiftEvents: ShiftEvent[];
		shiftStart: ZonedDateTime;
	};
	const { shiftEvents, shiftStart }: Props = $props();

	const format = $derived(settings.format);

	const eventKindIcons: Record<ShiftEvent["eventType"], string> = {
		["incident"]: mdiFire,
		["alert"]: mdiPhoneAlert,
	};

	const severityColors: Record<string, string> = {
		critical: "bg-error-500 text-white",
		high: "bg-error-300 text-black",
		medium: "bg-warning-300 text-black",
		low: "bg-info-300 text-black",
		undefined: "bg-surface-300 text-black"
	};

	const statusIcons: Record<string, string> = {
		active: mdiAlertCircle,
		resolved: mdiCheckCircle,
		acknowledged: mdiClockOutline,
		undefined: mdiClockOutline
	};

	const getRelativeTime = (date: Date) => {
		return formatDistanceToNow(date, { addSuffix: true });
	};
</script>

<div class="h-10 flex w-full gap-4 items-center px-2">
	<Header title="Shift Events" classes={{ root: "w-full", title: "text-xl", container: "flex-1" }}>
		<!--div slot="actions">
			<Button
				color="primary"
				variant="fill"
				on:click={() => {}}
			>
				Filters <Icon data={mdiPlus} />
			</Button>
		</div-->
	</Header>
</div>

<div class="flex-1 min-h-0 flex flex-col gap-4 overflow-y-auto bg-surface-200 p-3">
	{#each shiftEvents as ev}
		{@render eventListItem(ev)}
	{/each}
</div>

{#snippet eventListItem(ev: ShiftEvent)}
	{@const occurredAt = ev.timestamp.toDate()}
	{@const relativeTime = getRelativeTime(occurredAt)}
	{@const severityClass = severityColors[ev.severity || 'undefined']}
	{@const statusIcon = statusIcons[ev.status || 'undefined']}
	
	<div class="grid grid-cols-[100px_auto_minmax(0,1fr)] gap-2 place-items-center border rounded-md p-3 bg-surface-100 shadow-sm hover:shadow-md transition-shadow">
		<div class="justify-self-start flex flex-col items-start">
			<span class="text-sm font-medium">
				{format(occurredAt, PeriodType.Day)}
			</span>
			<span class="text-xs text-surface-600 flex items-center gap-1">
				<Icon data={mdiClockOutline} size="14px" />
				{relativeTime}
			</span>
		</div>

		<div class="items-center static z-10">
			<Icon
				data={eventKindIcons[ev.eventType]}
				classes={{ root: `${ev.eventType === 'incident' ? 'bg-error-600' : 'bg-warning-600'} rounded-full p-2 w-auto h-10 text-white` }}
			/>
		</div>

		<div class="w-full justify-self-start grid grid-cols-[1fr_auto] items-start gap-2 px-2">
			<div class="flex flex-col">
				<div class="font-medium">
					{ev.title || `${ev.eventType.charAt(0).toUpperCase() + ev.eventType.slice(1)} ${ev.id.substring(0, 8)}`}
				</div>
				{#if ev.source}
					<div class="text-xs text-surface-600">
						Source: {ev.source}
					</div>
				{/if}
			</div>
			<div class="flex items-center gap-2">
				{#if ev.severity}
					<Badge class={severityClass}>
						{ev.severity}
					</Badge>
				{/if}
				{#if ev.status}
					<div class="flex items-center gap-1 text-xs">
						<Icon data={statusIcon} size="16px" />
						{ev.status}
					</div>
				{/if}
			</div>
		</div>

		{#if ev.description || ev.notes}
			<div class="row-start-2 col-span-3 overflow-y-auto max-h-20 border rounded p-2 w-full bg-surface-50 text-sm">
				{#if ev.description}
					<div class="font-medium mb-1">{ev.description}</div>
				{/if}
				{#if ev.notes}
					<div class="text-surface-700">{ev.notes}</div>
				{/if}
			</div>
		{/if}
	</div>
{/snippet}
