<script lang="ts">
	import { mdiCalendar } from "@mdi/js";
	import { formatDate } from "date-fns";
	import Icon from "$components/icon/Icon.svelte";
	import { Tooltip } from "svelte-ux";
	import { getEventTimeIcon } from "./events";
	
	type Props = {
		timestamp: string;
	};
	const { timestamp }: Props = $props();
	
	const date = $derived(new Date(timestamp));
	const humanDate = $derived(formatDate(date, 'MMM d'));
	const humanTime = $derived(formatDate(date, 'h:mm a'));
	const timeIcon = $derived(getEventTimeIcon(date));
</script>

<div class="flex flex-col gap-1 justify-between w-full items-start">
	<span class="text-sm flex items-center gap-1">
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