<script lang="ts">
	import * as Empty from "$components/ui/empty";
	import SystemMapCanvas from "$features/systems/views/map/SystemMapCanvas.svelte";
	import { useSituationController } from "../controller.svelte";

	const controller = useSituationController();
	const mapEntityId = $derived(controller.situation?.attributes.knowledgeEntityId ?? "");
</script>

{#if controller.situation}
	<section class="flex min-h-0 min-w-0 flex-1 flex-col">
		<header class="flex flex-col gap-2 p-4">
			<h1 class="text-[28px] leading-9 font-semibold">Impact Scope</h1>
			<p class="max-w-[75ch] text-sm text-muted-foreground">
				System relationships around this Situation. Displayed entities are not necessarily confirmed
				affected.
			</p>
		</header>
		{#if mapEntityId}
			<SystemMapCanvas defaultFocus={() => mapEntityId} />
		{:else}
			<Empty.Root>
				<Empty.Header>
					<Empty.Title>Map context unavailable</Empty.Title>
					<Empty.Description>This Situation has no knowledge entity to focus on.</Empty.Description>
				</Empty.Header>
			</Empty.Root>
		{/if}
	</section>
{/if}
