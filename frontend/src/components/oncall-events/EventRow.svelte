<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	import { mdiPin, mdiPinOutline, mdiPhoneAlert, mdiFire, mdiChatQuestion, mdiChatPlus } from "@mdi/js";
	import { mdiCalendar, mdiClockOutline, mdiSleepOff, mdiWeatherSunset } from "@mdi/js";
	import { Button, Lazy, Tooltip } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { isBusinessHours, isNightHours } from "$features/oncall/lib/utils";
	import { formatDate } from "date-fns";

	type Props = {
		event: OncallEvent;
		annotations?: OncallAnnotation[];
		annotatableRosterIds?: string[];
		editAnnotation?: (anno?: OncallAnnotation) => void;
		pinned?: boolean;
		togglePinned?: () => void;
		loadingId?: string;
	}
	const { event, annotations = [], annotatableRosterIds = [], editAnnotation, pinned, togglePinned, loadingId }: Props = $props();

	const attrs = $derived(event.attributes);

	const date = $derived(new Date(attrs.timestamp));
	// const humanDate = $derived(formatDate(date, 'EEE, MMM d'));
	const humanDate = $derived(formatDate(date, 'MMM d'));
	const humanTime = $derived(formatDate(date, 'h:mm a'));
	const isOutsideBusinessHours = $derived(!isBusinessHours(date.getHours()));
	const isNightTime = $derived(isNightHours(date.getHours()));

	const rosterIdsWithAnnotations = $derived(new Set(annotations.map(a => a.attributes.roster.id)));
	const showAnnotationButton = $derived(annotatableRosterIds.some(id => !rosterIdsWithAnnotations.has(id)));
	const loading = $derived(!!loadingId && loadingId === event.id);
	const disabled = $derived(!!loadingId && loadingId !== event.id);

	const kindIcon = $derived.by(() => {
		switch (attrs.kind) {
		case "incident": return {icon: mdiFire, color: "text-danger-900/50"};
		case "alert": return {icon: mdiPhoneAlert, color: "text-warning-700/50"};
		default: return {icon: mdiChatQuestion, color: "text-surface-content/40"};
		}
	});

	const timeIcon = $derived.by(() => {
		if (isNightTime) return {tooltip: "Night hours (10pm-6am)", icon: mdiSleepOff, color: "text-danger-600"};
		if (isOutsideBusinessHours) return {tooltip: "Outside business hours (9am-5pm)", icon: mdiWeatherSunset, color: "text-warning-500"};
		return {tooltip: "", icon: mdiClockOutline, color: "text-surface-content/70"};
	});
</script>

<Lazy height="70px" class="group grid grid-cols-[80px_auto_auto] gap-2 place-items-center border p-2 bg-neutral-900/40 border-neutral-content/10 shadow-sm hover:shadow-md transition-shadow">
	<div class="flex flex-col gap-1 justify-between w-full items-start">
		<span class="text-sm font- flex items-center gap-1">
			<Icon data={mdiCalendar} size="16px" />
			{humanDate}
		</span>

		<Tooltip title={timeIcon.tooltip} placement="right" classes={{content: "flex items-center gap-1"}}>
			<span class="{timeIcon.color} leading-none">
				<Icon data={timeIcon.icon} size="16px" />
			</span>
					
			<span class="text-sm text-surface-700 inline-block align-middle">{humanTime}</span>
		</Tooltip>
	</div>

	<div class="flex flex-col gap-1 w-full h-full justify-center items-start">
		<div class="flex gap-1 items-center">
			<Icon data={kindIcon.icon} classes={{ root: `rounded-full size-4 w-auto ${kindIcon.color}` }} />
			<span class="text-xs uppercase font-normal text-surface-content/50">{attrs.kind}</span>
		</div>
		<div class="font- flex items-center leading-none">{attrs.title}</div>
	</div>

	<div class="flex w-full items-center justify-end">
		<div class="flex flex-1 justify-end gap-2">
			{#each annotations as anno}
				<div class="overflow-y-auto w-full h-full border rounded p-2 bg-neutral-700/70 text-sm flex items-center">
					<div class="text-neutral-content">{anno.attributes.notes}</div>
				</div>
			{:else} 
				{#if editAnnotation && showAnnotationButton}
					<div class="hidden group-hover:inline w-fit h-full">
						<Button classes={{root: "w-full h-full items-center"}} {loading} {disabled} on:click={() => editAnnotation()}>
							Annotate
							<Icon data={mdiChatPlus} />
						</Button>
					</div>
				{/if}
			{/each}
		</div>

		{#if !!togglePinned}
			<div class="self-end">
				<Button iconOnly icon={pinned ? mdiPin : mdiPinOutline} {loading} {disabled} on:click={togglePinned} />
			</div>
		{/if}
	</div>
</Lazy>