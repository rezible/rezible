<script lang="ts">
	import * as Tooltip from "$components/ui/tooltip";
	import type { BriefServiceView } from "$features/situations/lib/brief";
	import { catalogueProposalPath } from "$features/system/lib/catalogue/links";

	type Props = { service: BriefServiceView };
	const { service }: Props = $props();
</script>

<div class="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
	<span class="text-sm font-medium wrap-anywhere">{service.name}</span>
	{#each service.unconfirmedMatches as match (match.key)}
		<Tooltip.Root>
			<Tooltip.Trigger>
				{#snippet child({ props })}
					<a
						{...props}
						href={catalogueProposalPath(match.proposalId)}
						class="text-xs text-muted-foreground underline decoration-dotted underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
					>
						{match.label}
					</a>
				{/snippet}
			</Tooltip.Trigger>
			<Tooltip.Content class="max-w-80">
				<p>{match.rationale}</p>
			</Tooltip.Content>
		</Tooltip.Root>
	{/each}
</div>
