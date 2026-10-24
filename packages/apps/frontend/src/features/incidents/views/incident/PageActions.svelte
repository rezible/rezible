<script lang="ts">
	import RiMoreLine from "remixicon-svelte/icons/more-line";

	import { Button } from "$components/ui/button";
	import * as DropdownMenu from "$components/ui/dropdown-menu";
	import type { ResponseAction } from "$features/incidents/lib/response";

	import type { IncidentViewController } from "./controller.svelte";
	import type { IncidentResponseController } from "./response.svelte";

	type Props = { controller: IncidentViewController; response: IncidentResponseController };
	const { controller, response }: Props = $props();

	const actionLabels: Record<ResponseAction, string> = {
		mitigated: "Mitigated",
		mitigationFailed: "Mitigation failed",
		resolve: "Resolve",
	};
</script>

{#each response.actions as action (action)}
	<Button
		size="sm"
		variant={action === "resolve" ? "default" : "outline"}
		onclick={() => response.openResponseAction(action)}
	>
		{actionLabels[action]}
	</Button>
{/each}

<Button variant="outline" size="sm" onclick={controller.share}>Share</Button>

<DropdownMenu.Root>
	<DropdownMenu.Trigger>
		{#snippet child({ props })}
			<Button {...props} size="icon-sm" variant="outline" aria-label="More incident actions">
				<RiMoreLine />
			</Button>
		{/snippet}
	</DropdownMenu.Trigger>
	<DropdownMenu.Content align="end" class="w-52">
		<DropdownMenu.Group>
			<DropdownMenu.Item onSelect={response.openEditDetails}>Edit details…</DropdownMenu.Item>
			{#if response.enabled}
				<DropdownMenu.Item onSelect={response.openRecordMilestone}>
					Record milestone…
				</DropdownMenu.Item>
			{/if}
			<DropdownMenu.Item onSelect={response.openLinkSituation}>Link situation…</DropdownMenu.Item>
		</DropdownMenu.Group>
		{#if response.isAdmin}
			<DropdownMenu.Separator />
			<DropdownMenu.Item class="text-destructive" onSelect={response.openArchive}>
				Archive incident…
			</DropdownMenu.Item>
		{/if}
	</DropdownMenu.Content>
</DropdownMenu.Root>
