<script lang="ts">
	import { statusToneTextClass, type StatusPresentation } from "./status";
	import { cn } from "$lib/utils";
	import { badgeVariants } from "$components/ui/badge";
	import * as Tooltip from "$components/ui/tooltip";

	type Props = {
		status: StatusPresentation;
		/** "pill" = filled rounded badge; "inline" = icon + coloured text, no fill. */
		variant?: "pill" | "inline";
		class?: string;
	};

	const { status, variant = "pill", class: className }: Props = $props();

	const classes = $derived.by(() => {
		if (variant === "inline") {
			return cn(
				"inline-flex items-center gap-1.5 whitespace-nowrap rounded-sm text-xs font-medium focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring [&>svg]:size-4 [&>svg]:shrink-0",
				statusToneTextClass[status.tone],
				className
			);
		}
		return cn(badgeVariants({ variant: status.tone }), className);
	});

	const accessibleLabel = $derived(
		status.description ? `${status.label}: ${status.description}` : undefined
	);
</script>

{#snippet badge(props: Record<string, unknown> = {})}
	<span {...props} data-slot="status-badge" class={classes} aria-label={accessibleLabel}>
		<status.icon aria-hidden="true" />
		{status.label}
	</span>
{/snippet}

{#if status.description}
	<Tooltip.Root>
		<Tooltip.Trigger>
			{#snippet child({ props })}
				{@render badge({ ...props, tabindex: 0 })}
			{/snippet}
		</Tooltip.Trigger>
		<Tooltip.Content>{status.description}</Tooltip.Content>
	</Tooltip.Root>
{:else}
	{@render badge()}
{/if}
