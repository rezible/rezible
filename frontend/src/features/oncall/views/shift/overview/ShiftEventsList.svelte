<script lang="ts">
	import { mdiPhoneAlert, mdiFire, mdiCalendarClock, mdiSleepOff, mdiWeatherSunset, mdiFilter, mdiChatQuestion } from "@mdi/js";
	import { Icon, Header, Tooltip, Button } from "svelte-ux";
	import { isBusinessHours, isNightHours } from "$features/oncall/lib/utils";
	import { format as formatDate } from "date-fns";
	import type { OncallEvent, OncallEventAnnotation } from "$lib/api";

	type Props = {
		events: OncallEvent[];
		annotations: OncallEventAnnotation[];
	};
	const { events, annotations }: Props = $props();

	const getEventKindIcon = (kind: string) => {
		switch (kind) {
		case "incident": return mdiFire;
		case "alert": return mdiPhoneAlert;
		}
		return mdiChatQuestion;
	};

	const humanReadableDate = (date: Date) => {
		return formatDate(date, 'EEE, MMM d');
	};

	const humanReadableTime = (date: Date) => {
		return formatDate(date, 'h:mm a');
	};

	const annotationsMap = $derived.by(() => {
		const m = new Map<string, OncallEventAnnotation>();
		annotations.forEach(a => {
			if (!a.attributes.event) return;
			m.set(a.attributes.event.id, a);
		})
		return m;
	});
</script>

<div class="flex flex-col gap-2 h-full border border-surface-content/10 rounded">
	<div class="h-fit p-2 pb-0">
		<Header title="Events" subheading="Showing All">
			<svelte:fragment slot="actions">
				<Button icon={mdiFilter} iconOnly />
			</svelte:fragment>
		</Header>
	</div>
	<div class="flex-1 flex flex-col gap-1 px-0 overflow-y-auto">
		{#each events as ev}
			{@render eventListItem(ev)}
		{/each}
	</div>
</div>

{#snippet eventListItem(ev: OncallEvent)}
	{@const occurredAt = new Date(ev.timestamp)}
	{@const humanDate = humanReadableDate(occurredAt)}
	{@const humanTime = humanReadableTime(occurredAt)}
	{@const isOutsideBusinessHours = !isBusinessHours(occurredAt.getHours())}
	{@const isNightTime = isNightHours(occurredAt.getHours())}
	{@const icon = getEventKindIcon(ev.kind)}
	{@const annotation = annotationsMap.get(ev.id)}
	
	<div class="grid grid-cols-[auto_minmax(0,1fr)_120px] gap-2 place-items-center border p-3 bg-neutral-900/40 border-neutral-content/10 shadow-sm hover:shadow-md transition-shadow">
		<div class="items-center static z-10">
			<Icon
				data={icon}
				classes={{ root: `${ev.kind === 'incident' ? 'bg-danger-900/50' : 'bg-warning-700/50'} rounded-full p-2 w-auto h-10 text-white` }}
			/>
		</div>

		<div class="w-full justify-self-start grid grid-cols-[1fr_auto] items-start gap-2 px-2">
			<div class="flex flex-col">
				<div class="font-medium">
					{ev.title || `${ev.kind.charAt(0).toUpperCase() + ev.kind.slice(1)} ${ev.id.substring(0, 8)}`}
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

		{#if annotation}
			<div class="row-start-2 col-span-3 overflow-y-auto max-h-20 border rounded p-2 w-full bg-neutral-700/70 text-sm">
				<div class="text-neutral-content">{annotation.attributes.notes}</div>
			</div>
		{/if}
	</div>
{/snippet}
