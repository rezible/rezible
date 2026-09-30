<script lang="ts">
	import * as Tooltip from "$components/ui/tooltip";
	import { formatTime } from "$lib/time";
	import { cn } from "$lib/utils";

	type Props = {
		value: string | null | undefined;
		format?: "relative" | "clock" | "absolute";
		class?: string;
	};

	const { value, format = "relative", class: className }: Props = $props();

	const formatted = $derived(formatTime(value));
	const label = $derived.by(() => {
		if (format === "clock") {
			return formatted.clock;
		}
		if (format === "absolute") {
			return formatted.absolute;
		}
		return formatted.relative;
	});
	const showTooltip = $derived(format !== "absolute" && !!formatted.iso);
</script>

{#if showTooltip}
	<Tooltip.Root>
		<Tooltip.Trigger
			aria-label={formatted.absolute}
			class={cn(
				"cursor-default whitespace-nowrap rounded-sm tabular-nums focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
				className
			)}
		>
			<time datetime={formatted.iso}>{label}</time>
		</Tooltip.Trigger>
		<Tooltip.Content>{formatted.absolute}</Tooltip.Content>
	</Tooltip.Root>
{:else}
	<time datetime={formatted.iso} class={cn("tabular-nums", className)}>{label}</time>
{/if}
