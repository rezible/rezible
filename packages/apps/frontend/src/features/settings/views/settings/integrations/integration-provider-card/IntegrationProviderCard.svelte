<script lang="ts">
	import RiArrowRightLine from "remixicon-svelte/icons/arrow-right-line";

	import { resolve } from "$app/paths";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";

	import type { IntegrationProviderSummary } from "$features/settings/lib/integrationsController.svelte";

	type Props = {
		provider: IntegrationProviderSummary;
	};
	const { provider }: Props = $props();

	const installed = $derived(provider.installations.length > 0);
	const actionLabel = $derived(installed ? "Manage" : "Set up");
	const href = $derived(resolve("/settings/integrations/[provider]", { provider: provider.name }));
</script>

<!-- The action link covers the whole card, so the card is one large click target. -->
<Card.Root
	class="relative h-full transition-colors focus-within:ring-2 focus-within:ring-ring hover:border-foreground/25"
>
	<Card.Header>
		<div class="flex min-w-0 items-start gap-3">
			<div class="grid size-11 shrink-0 place-items-center border bg-muted">
				<provider.icon class="size-6" />
			</div>
			<div class="flex min-w-0 flex-col gap-2">
				<h3 class="text-base leading-tight font-medium">{provider.displayName}</h3>
				{#if provider.capabilityPills.length > 0}
					<div class="flex flex-wrap gap-1">
						{#each provider.capabilityPills as pill (pill)}
							<Badge variant="outline" class="font-normal">{pill}</Badge>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	</Card.Header>

	<Card.Content>
		<p class="line-clamp-2 text-sm text-muted-foreground">{provider.description}</p>
	</Card.Content>

	<Card.Footer class="mt-auto justify-between gap-3">
		<!-- Connection health from data sync will share this status line. -->
		{#if provider.status}
			<span class="flex min-w-0 items-center gap-2 text-xs">
				<span
					class="size-2 shrink-0 rounded-full bg-status-success-foreground"
					aria-hidden="true"
				></span>
				<span class="truncate">{provider.status}</span>
			</span>
		{:else}
			<span></span>
		{/if}
		<Button
			size="sm"
			variant={installed ? "outline" : "default"}
			{href}
			aria-label="{actionLabel} {provider.displayName}"
			class="after:absolute after:inset-0"
		>
			{actionLabel}
			{#if !installed}
				<RiArrowRightLine data-icon="inline-end" />
			{/if}
		</Button>
	</Card.Footer>
</Card.Root>
