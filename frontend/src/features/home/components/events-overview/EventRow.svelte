<script lang="ts">
	import type { OncallEvent, OncallEventAnnotation } from "$lib/api";
	import { format as formatDate } from "date-fns";
	import { mdiAlert, mdiCalendarClock, mdiCircleMedium, mdiClock, mdiFire, mdiSleepOff, mdiWeatherSunset } from "@mdi/js";
	import { Button, Checkbox, Icon, Tooltip } from "svelte-ux";
	import { isBusinessHours, isNightHours } from "$features/oncall/lib/utils";

	type Props = {
		event: OncallEvent;
		checked: boolean;
		onToggleChecked: () => void;
		onOpenAnnotateDialog: () => void;
	}
	let { event, checked, onToggleChecked, onOpenAnnotateDialog }: Props = $props();

	const eventIcon = $derived.by(() => {
		switch (event.kind) {
			case "incident": return mdiFire;
			case "alert": return mdiAlert;
		}
		return mdiCircleMedium;
	});


	const date = $derived(new Date(event.timestamp));
	const humanDate = $derived(formatDate(date, 'EEE, MMM d'))
	const humanTime = $derived(formatDate(date, 'h:mm a'));
	const isOutsideBusinessHours = $derived(!isBusinessHours(date.getHours()));
	const isNightTime = $derived(isNightHours(date.getHours()));

	let hovering = $state(false);
	const buttonVariant = $derived(hovering ? "fill" : "fill-light");
</script>

<div class="grid grid-cols-subgrid col-span-full hover:bg-surface-100/50 h-16 p-2" onpointerenter={() => (hovering=true)} onpointerleave={()=>(hovering=false)}>
	<div class="grid place-self-center">
		<Checkbox {checked} on:change={onToggleChecked} />
	</div>
	<div class="grid place-items-center">
		<Icon data={eventIcon} />
	</div>
	<div class="grid items-center">
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
	</div>
	<div class="flex items-center">
		<span>{event.title}</span>
	</div>
	<div class="flex items-center">
		<span></span>
	</div>
	<div class="flex items-center justify-end">
		<Button variant={buttonVariant} size="sm" on:click={onOpenAnnotateDialog}>Edit</Button>
	</div>
</div>