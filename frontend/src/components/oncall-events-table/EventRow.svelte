<script lang="ts">
	import type { OncallAnnotation, OncallEvent } from "$lib/api";
	import { format as formatDate } from "date-fns";
	import { mdiAlert, mdiCalendarClock, mdiCircleMedium, mdiFire, mdiPlus, mdiSleepOff, mdiWeatherSunset } from "@mdi/js";
	import { Button, Icon, Tooltip } from "svelte-ux";
	import { isBusinessHours, isNightHours } from "$features/oncall/lib/utils";

	type Props = {
		event: OncallEvent;
		allowAnnotationRosters: string[];
		onOpenAnnotateDialog: (anno?: OncallAnnotation) => void;
	}
	let { event, allowAnnotationRosters, onOpenAnnotateDialog }: Props = $props();

	const attrs = $derived(event.attributes);

	const eventIcon = $derived.by(() => {
		switch (attrs.kind) {
			case "incident": return mdiFire;
			case "alert": return mdiAlert;
		}
		return mdiCircleMedium;
	});

	const annoRosters = $derived(new Set(event.attributes.annotations.map(a => a.attributes.rosterId)));
	const needsAnnotation = $derived(allowAnnotationRosters.some(id => !annoRosters.has(id)));

	const date = $derived(new Date(attrs.timestamp));
	const humanDate = $derived(formatDate(date, 'EEE, MMM d'))
	const humanTime = $derived(formatDate(date, 'h:mm a'));
	const isOutsideBusinessHours = $derived(!isBusinessHours(date.getHours()));
	const isNightTime = $derived(isNightHours(date.getHours()));

	let hovering = $state(false);
</script>

<div class="grid grid-cols-subgrid col-span-full hover:bg-surface-100/50 h-16 p-2" onpointerenter={() => (hovering=true)} onpointerleave={()=>(hovering=false)}>
	<div class="grid place-items-center">
		<Icon data={eventIcon} />
	</div>
	<div class="grid items-center">
		{@render eventTime()}
	</div>
	<div class="flex items-center">
		<span>{attrs.title}</span>
	</div>
	<div class="flex items-center justify-end">
		{#each event.attributes.annotations as anno}
			<div class="border p-1">
				<span>{anno.attributes.notes}</span>
				{#if allowAnnotationRosters.includes(anno.attributes.rosterId)}
					<Button variant="fill-light" on:click={() => (onOpenAnnotateDialog(anno))}>
						<span class="flex items-center gap-2">
							Edit
						</span>
					</Button>
				{/if}
			</div>
		{/each}

		{#if needsAnnotation}
			<Button variant="fill-light" color={hovering ? "accent" : "default"} on:click={() => (onOpenAnnotateDialog())}>
				<span class="flex items-center gap-2">
					Annotate
					<Icon data={mdiPlus} />
				</span>
			</Button>
		{/if}
	</div>
</div>

{#snippet eventTime()}
	<div class="flex flex-col items-start">
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
{/snippet}
