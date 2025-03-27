<script lang="ts">
	import { mdiPhoneAlert, mdiFire, mdiCheckCircle, mdiClockOutline, mdiAlertCircle, mdiCalendarClock, mdiMoonWaxingCrescent, mdiWeatherNight, mdiTimerOffOutline, mdiSleep, mdiSleepOff, mdiWeatherSunset, mdiFilter } from "@mdi/js";
	import { Icon, Header, Badge, Tooltip, Button } from "svelte-ux";
	import { settings } from "$lib/settings.svelte";
	import type { ShiftEvent } from "$src/features/oncall/lib/utils";
	import { isBusinessHours, isNightHours } from "$src/features/oncall/lib/utils";
	import { PeriodType } from "@layerstack/utils";
	import { formatDistanceToNow, format as formatDate, isToday, isYesterday, isTomorrow } from "date-fns";
	import type { ZonedDateTime } from "@internationalized/date";

	type Props = {
		shiftEvents: ShiftEvent[];
	};
	const { shiftEvents }: Props = $props();

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

	const getHumanReadableDate = (date: Date) => {
		return formatDate(date, 'EEE, MMM d');
	};

	const getHumanReadableTime = (date: Date) => {
		return formatDate(date, 'h:mm a');
	};
</script>

<div class="flex flex-col gap-2 h-full border bg-surface-100/40 border-surface-content/10 rounded">
	<div class="h-fit p-2 pb-0">
		<Header title="Events" subheading="Showing All">
			<svelte:fragment slot="actions">
				<Button icon={mdiFilter} iconOnly />
			</svelte:fragment>
		</Header>
	</div>
	<div class="flex-1 flex flex-col gap-1 px-0 overflow-y-auto">
		{#each shiftEvents as ev}
			{@render eventListItem(ev)}
		{/each}
	</div>
</div>

{#snippet eventListItem(ev: ShiftEvent)}
	{@const occurredAt = ev.timestamp.toDate()}
	{@const humanDate = getHumanReadableDate(occurredAt)}
	{@const humanTime = getHumanReadableTime(occurredAt)}
	{@const severityClass = severityColors[ev.severity || 'undefined']}
	{@const statusIcon = statusIcons[ev.status || 'undefined']}
	{@const isOutsideBusinessHours = !isBusinessHours(ev.timestamp.hour)}
	{@const isNightTime = isNightHours(ev.timestamp.hour)}
	
	<div class="grid grid-cols-[auto_minmax(0,1fr)_120px] gap-2 place-items-center border p-3 bg-neutral-900/40 border-neutral-content/10 shadow-sm hover:shadow-md transition-shadow">
		<div class="items-center static z-10">
			<Icon
				data={eventKindIcons[ev.eventType]}
				classes={{ root: `${ev.eventType === 'incident' ? 'bg-danger-900/50' : 'bg-warning-700/50'} rounded-full p-2 w-auto h-10 text-white` }}
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
				{#if ev.description}
					<div class="text-xs text-surface-600">
						{ev.description}
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

		<div class="justify-self-end flex flex-col items-start">
			<div class="flex flex-col">
				<span class="text-sm font-medium flex items-center gap-1">
					<Icon data={mdiCalendarClock} size="14px" />
					{humanDate}
				</span>
				<div class="flex items-center gap-1">
					{#if isNightTime}
						<Tooltip title="Night hours (10pm-6am)" placement="right">
							<span class="text-danger-600">
								<Icon data={mdiSleepOff} size="16px" />
							</span>
						</Tooltip>
					{:else if isOutsideBusinessHours}
						<Tooltip title="Outside business hours (9am-5pm)" placement="right">
							<span class="text-warning-600">
								<Icon data={mdiWeatherSunset} size="16px" />
							</span>
						</Tooltip>
					{/if}
					<span class="text-sm text-surface-700 self-end">
						{humanTime}
					</span>
				</div>
			</div>
		</div>

		{#if ev.annotation}
			<div class="row-start-2 col-span-3 overflow-y-auto max-h-20 border rounded p-2 w-full bg-neutral-700/70 text-sm">
				<div class="text-neutral-content">{ev.annotation}</div>
			</div>
		{/if}
	</div>
{/snippet}
