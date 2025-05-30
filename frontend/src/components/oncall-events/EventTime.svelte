<script lang="ts">
	import { format as formatDate } from "date-fns";
	import { isBusinessHours, isNightHours } from "$features/oncall/lib/utils";
	import { mdiAlert, mdiCalendarClock, mdiSleepOff, mdiWeatherSunset } from "@mdi/js";
	import { Tooltip } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";

	type Props = {
		timestamp: string;
	};
	const { timestamp }: Props = $props();

	const date = $derived(new Date(timestamp));
	// const humanDate = $derived(formatDate(date, 'EEE, MMM d'))
	const humanDate = $derived(formatDate(date, 'MMM d'))
	const humanTime = $derived(formatDate(date, 'h:mm a'));
	const isOutsideBusinessHours = $derived(!isBusinessHours(date.getHours()));
	const isNightTime = $derived(isNightHours(date.getHours()));
</script>

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