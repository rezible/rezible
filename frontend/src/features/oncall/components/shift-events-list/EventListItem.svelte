<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	import { mdiPhoneAlert, mdiFire, mdiCalendarClock, mdiSleepOff, mdiWeatherSunset, mdiFilter, mdiChatQuestion } from "@mdi/js";
	import { Icon, Header, Tooltip, Button } from "svelte-ux";
	import { isBusinessHours, isNightHours } from "$features/oncall/lib/utils";
	import { format as formatDate } from "date-fns";

	type Props = {
		event: OncallEvent;
		onEditAnnotation: () => void;
	}
	const { event, onEditAnnotation }: Props = $props();

	const getEventKindIcon = (kind: string) => {
		switch (kind) {
		case "incident": return mdiFire;
		case "alert": return mdiPhoneAlert;
		}
		return mdiChatQuestion;
	};

	const humanReadableDate = (date: Date) => formatDate(date, 'EEE, MMM d');
	const humanReadableTime = (date: Date) => formatDate(date, 'h:mm a');

	const attrs = $derived(event.attributes)
	const occurredAt = $derived(new Date(attrs.timestamp))
	const humanDate = $derived(humanReadableDate(occurredAt))
	const humanTime = $derived(humanReadableTime(occurredAt))
	const isOutsideBusinessHours = $derived(!isBusinessHours(occurredAt.getHours()))
	const isNightTime = $derived(isNightHours(occurredAt.getHours()))
	const icon = $derived(getEventKindIcon(attrs.kind))

	const annotation = $derived(attrs.annotations.at(0));
</script>

<div class="group grid grid-cols-[auto_minmax(0,1fr)_120px] gap-2 place-items-center border p-3 bg-neutral-900/40 border-neutral-content/10 shadow-sm hover:shadow-md transition-shadow">
	<div class="items-center static z-10">
		<Icon
			data={icon}
			classes={{ root: `${attrs.kind === 'incident' ? 'bg-danger-900/50' : 'bg-warning-700/50'} rounded-full p-2 w-auto h-10 text-white` }}
		/>
	</div>

	<div class="w-full justify-self-start grid grid-cols-[1fr_auto] items-start gap-2 px-2">
		<div class="flex flex-col h-full">
			<div class="font-medium h-full flex items-center">
				{attrs.title || `${attrs.kind.charAt(0).toUpperCase() + attrs.kind.slice(1)} ${event.id.substring(0, 8)}`}
			</div>
		</div>

		{#if annotation}
			<div class="overflow-y-auto max-h-20 border rounded p-2 w-full bg-neutral-700/70 text-sm">
				<div class="text-neutral-content">{annotation.attributes.notes}</div>
			</div>
		{:else}
			<div class="border hidden group-hover:inline">
				<Button on:click={onEditAnnotation}>Add Annotation</Button>
			</div>
		{/if}
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
</div>