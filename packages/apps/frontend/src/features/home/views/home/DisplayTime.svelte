<script lang="ts">
	import * as Tooltip from "$components/ui/tooltip";
	import { timestamp } from "./model";
	import { cn } from "$lib/utils";

	type Props = { value?: string; clock?: boolean; class?: string };
	let { value, clock = false, class: className = "" }: Props = $props();

	const time = $derived(timestamp(value));
</script>

<Tooltip.Root>
	<Tooltip.Trigger
		class={cn(
			"self-start rounded-sm whitespace-nowrap text-xs text-muted-foreground focus-visible:outline-2 focus-visible:outline-ring",
			className
		)}
		aria-label={time.full}
	>
		<time datetime={time.iso}>{clock ? time.clock : time.relative}</time>
	</Tooltip.Trigger>
	<Tooltip.Content>{time.full}</Tooltip.Content>
</Tooltip.Root>
