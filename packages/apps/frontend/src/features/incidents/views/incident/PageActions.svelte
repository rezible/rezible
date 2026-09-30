<script lang="ts">
	import { resolve } from "$app/paths";
	import { Button } from "$components/ui/button";
	import RiExternalLinkLine from "remixicon-svelte/icons/external-link-line";
	import type { IncidentViewController } from "./controller.svelte";
	import * as Popover from "$components/ui/popover";

	type Props = { controller: IncidentViewController };
	let { controller }: Props = $props();

	const situations = $derived(controller.situations);
</script>

<div class="flex shrink-0 items-center justify-end gap-1 sm:gap-2">
	{#if situations.length === 1}
		{@const situation = situations[0]}
		<a
			class="inline-flex items-center gap-2 rounded-md border border-border px-3 py-2 text-sm"
			href={resolve("/situations/[id]/[[view=situationView]]", { id: situation.id })}
		>
			<span class="max-w-40 truncate">{situation.title}</span>
			<RiExternalLinkLine aria-hidden="true" />
		</a>
	{:else if situations.length > 1}
		<Popover.Root>
			<Popover.Trigger>
				{#snippet child({ props })}
					<Button {...props} variant="outline" size="sm" class="shrink-0">
						{situations.length} situations
					</Button>
				{/snippet}
			</Popover.Trigger>
			<Popover.Content align="end" class="flex max-w-[calc(100vw-2rem)] flex-col gap-2">
				{#each situations as situation (situation.id)}
					<a
						class="flex items-center justify-between gap-3 rounded-md px-2 py-1 text-sm hover:bg-muted"
						href={resolve("/situations/[id]/[[view=situationView]]", { id: situation.id })}
					>
						{situation.title}<RiExternalLinkLine aria-hidden="true" />
					</a>
				{/each}
			</Popover.Content>
		</Popover.Root>
	{/if}
	<Button variant="outline" size="sm" onclick={controller.share}>Share</Button>
</div>
