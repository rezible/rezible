<script lang="ts">
	import RiArrowRightLine from "remixicon-svelte/icons/arrow-right-line";
	import RiCheckboxCircleFill from "remixicon-svelte/icons/checkbox-circle-fill";

	import { resolve } from "$app/paths";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";

	import type { IntegrationProviderSummary } from "$features/settings/lib/integrationsController.svelte";

	type Props = {
		provider: IntegrationProviderSummary;
	};
	const { provider }: Props = $props();

	// Several apps can share a connection name, such as both Slack apps in one workspace.
	const connectionNames = $derived(
		Array.from(new Set(provider.installations.map((installation) => installation.attributes.displayName)))
	);
	const installed = $derived(connectionNames.length > 0);
	const installedLabel = $derived.by(() => {
		const [first, ...others] = connectionNames;
		if (others.length === 0) {
			return first;
		}
		return `${first} +${others.length} more`;
	});
	const href = $derived(resolve("/settings/integrations/[provider]", { provider: provider.name }));
</script>

<Card.Root class="h-full">
	<Card.Header>
		<div class="flex min-w-0 items-start gap-3">
			<div class="grid size-10 shrink-0 place-items-center border bg-muted">
				<provider.icon class="size-5" />
			</div>
			<div class="flex min-w-0 flex-col gap-1">
				<Card.Title class="truncate">{provider.displayName}</Card.Title>
				<Card.Description class="line-clamp-2">{provider.description}</Card.Description>
			</div>
		</div>
	</Card.Header>

	<Card.Footer class="mt-auto justify-between gap-2">
		{#if installed}
			<span class="flex min-w-0 items-center gap-1.5 text-xs text-muted-foreground">
				<RiCheckboxCircleFill class="size-4 shrink-0 text-primary" />
				<span class="truncate">{installedLabel}</span>
			</span>
			<Button size="sm" variant="outline" {href}>Manage</Button>
		{:else}
			<span class="text-xs text-muted-foreground">Not installed</span>
			<Button size="sm" {href}>
				Install
				<RiArrowRightLine data-icon="inline-end" />
			</Button>
		{/if}
	</Card.Footer>
</Card.Root>
